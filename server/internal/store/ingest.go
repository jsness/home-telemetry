package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"home-telemetry/server/internal/types"
)

func (s Stores) InsertIngest(ctx context.Context, nodeID string, ts time.Time, metrics []types.MetricRow, tags map[string]string) error {
	metaBytes, err := encodeMap(tags)
	if err != nil {
		return err
	}

	batch := &pgx.Batch{}
	batch.Queue(
		"INSERT INTO nodes (id, name, last_seen, meta) VALUES ($1, $2, $3, $4) " +
			"ON CONFLICT (id) DO UPDATE SET last_seen = EXCLUDED.last_seen, meta = EXCLUDED.meta",
		nodeID, nodeID, ts, metaBytes,
	)

	for _, m := range metrics {
		labelsBytes, err := encodeMap(m.Labels)
		if err != nil {
			return err
		}
		batch.Queue(
			"INSERT INTO metrics (time, node_id, metric, value, labels) VALUES ($1, $2, $3, $4, $5)",
			m.Time, nodeID, m.Metric, m.Value, labelsBytes,
		)
	}

	br := s.Pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func encodeMap(m map[string]string) ([]byte, error) {
	if len(m) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}
