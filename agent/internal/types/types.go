package types

import "time"

type IngestPayload struct {
	NodeID    string            `json:"node_id"`
	Timestamp string            `json:"timestamp"`
	CPU       *CPUMetrics       `json:"cpu,omitempty"`
	GPUs      []GPUMetrics      `json:"gpus,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type CPUMetrics struct {
	TempC    float64 `json:"temp_c,omitempty"`
	UsagePct float64 `json:"usage_pct"`
	LoadAvg  float64 `json:"load_avg,omitempty"`
	FreqMHz  float64 `json:"freq_mhz,omitempty"`
	Cores    int     `json:"cores,omitempty"`
}

type GPUMetrics struct {
	Name      string  `json:"name"`
	TempC    float64 `json:"temp_c,omitempty"`
	UsagePct  float64 `json:"usage_pct"`
	MemUsedMB float64 `json:"mem_used_mb"`
	PowerW    float64 `json:"power_w"`
}

func NewPayload(node string, cpu *CPUMetrics, gpus []GPUMetrics) IngestPayload {
	return IngestPayload{
		NodeID:    node,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		CPU:       cpu,
		GPUs:      gpus,
	}
}