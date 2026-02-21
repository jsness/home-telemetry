package store

import "time"

type Node struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	LastSeen time.Time         `json:"last_seen"`
	Meta     map[string]string `json:"meta"`
}

type MetricsQuery struct {
	NodeID string
	Metric string
	From   *time.Time
	To     *time.Time
	Limit  int
}
