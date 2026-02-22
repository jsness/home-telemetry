package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"home-telemetry/server/internal/auth"
	"home-telemetry/server/internal/config"
	"home-telemetry/server/internal/store"
	"home-telemetry/server/internal/types"
)

type Handler struct {
	cfg    config.Config
	stores store.Stores
	logger *log.Logger
}

func NewHandler(cfg config.Config, stores store.Stores, logger *log.Logger) *Handler {
	return &Handler{cfg: cfg, stores: stores, logger: logger}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(corsMiddleware(h.cfg.CORSOrigins))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

		r.Route("/api/v1", func(r chi.Router) {
		// Public read endpoints for UI
		r.Get("/nodes", h.handleNodes)
		r.Get("/metrics", h.handleMetrics)

		// Protected ingest endpoint
		r.With(auth.TokenMiddleware(h.cfg.AuthToken)).Post("/ingest", h.handleIngest)
	})
	return r
}

func corsMiddleware(allowed string) func(http.Handler) http.Handler {
	allowed = strings.TrimSpace(allowed)
	allowAll := allowed == "*" || allowed == ""
	allowedSet := map[string]struct{}{}
	if !allowAll {
		for _, o := range strings.Split(allowed, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				allowedSet[o] = struct{}{}
			}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowAll && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			} else if origin != "" {
				if _, ok := allowedSet[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
			}

			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// @Summary Ingest metrics
// @Description Accepts a single metrics payload.
// @Tags ingest
// @Accept json
// @Produce json
// @Param payload body types.IngestPayload true "Ingest payload"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /ingest [post]
func (h *Handler) handleIngest(w http.ResponseWriter, r *http.Request) {
	var payload types.IngestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if payload.NodeID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ts := time.Now().UTC()
	if payload.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, payload.Timestamp)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ts = parsed.UTC()
	}

	metrics := types.ToMetricRows(payload, ts)
	if err := h.stores.InsertIngest(r.Context(), payload.NodeID, ts, metrics, payload.Tags); err != nil {
		h.logger.Printf("ingest error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// @Summary List nodes
// @Tags nodes
// @Produce json
// @Success 200 {array} store.Node
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /nodes [get]
func (h *Handler) handleNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.stores.ListNodes(r.Context())
	if err != nil {
		h.logger.Printf("nodes error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(nodes)
}

// @Summary Query metrics
// @Tags metrics
// @Produce json
// @Param node_id query string true "Node ID"
// @Param metric query string false "Metric name"
// @Param from query string false "RFC3339 time"
// @Param to query string false "RFC3339 time"
// @Param limit query int false "Max rows (default 1000)"
// @Success 200 {object} map[string][]types.MetricRow
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /metrics [get]
func (h *Handler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	nodeID := q.Get("node_id")
	metric := q.Get("metric")
	fromStr := q.Get("from")
	toStr := q.Get("to")
	limitStr := q.Get("limit")

	var from *time.Time
	if fromStr != "" {
		parsed, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		pt := parsed.UTC()
		from = &pt
	}
	var to *time.Time
	if toStr != "" {
		parsed, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		pt := parsed.UTC()
		to = &pt
	}

	limit := 0
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
		}
	}

	rows, err := h.stores.QueryMetrics(r.Context(), store.MetricsQuery{
		NodeID: nodeID,
		Metric: metric,
		From:   from,
		To:     to,
		Limit:  limit,
	})
	if err != nil {
		h.logger.Printf("metrics error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"series": rows})
}
