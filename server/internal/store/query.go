package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"home-telemetry/server/internal/types"
)

func (s Stores) ListNodes(ctx context.Context) ([]Node, error) {
	rows, err := s.Pool.Query(ctx, "SELECT id, name, last_seen, meta FROM nodes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Node
	for rows.Next() {
		var n Node
		var metaBytes []byte
		if err := rows.Scan(&n.ID, &n.Name, &n.LastSeen, &metaBytes); err != nil {
			return nil, err
		}
		if len(metaBytes) > 0 {
			_ = json.Unmarshal(metaBytes, &n.Meta)
		}
		out = append(out, n)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (s Stores) QueryMetrics(ctx context.Context, q MetricsQuery) ([]types.MetricRow, error) {
	if q.NodeID == "" {
		return nil, fmt.Errorf("node_id required")
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}

	args := []any{q.NodeID}
	clauses := []string{"node_id = $1"}

	if q.Metric != "" {
		args = append(args, q.Metric)
		clauses = append(clauses, fmt.Sprintf("metric = $%d", len(args)))
	}
	if q.From != nil {
		args = append(args, *q.From)
		clauses = append(clauses, fmt.Sprintf("time >= $%d", len(args)))
	}
	if q.To != nil {
		args = append(args, *q.To)
		clauses = append(clauses, fmt.Sprintf("time <= $%d", len(args)))
	}

	args = append(args, limit)
	query := "SELECT time, metric, value, labels FROM metrics WHERE " + strings.Join(clauses, " AND ") +
		fmt.Sprintf(" ORDER BY time ASC LIMIT $%d", len(args))

	rows, err := s.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []types.MetricRow
	for rows.Next() {
		var r types.MetricRow
		var labelsBytes []byte
		if err := rows.Scan(&r.Time, &r.Metric, &r.Value, &labelsBytes); err != nil {
			return nil, err
		}
		if len(labelsBytes) > 0 {
			_ = json.Unmarshal(labelsBytes, &r.Labels)
		}
		out = append(out, r)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}
