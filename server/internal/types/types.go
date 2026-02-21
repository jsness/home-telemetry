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
	TempC    float64 `json:"temp_c"`
	UsagePct float64 `json:"usage_pct"`
	LoadAvg  float64 `json:"load_avg"`
	FreqMHz  float64 `json:"freq_mhz,omitempty"`
	Cores    int     `json:"cores,omitempty"`
}

type GPUMetrics struct {
	Name      string  `json:"name"`
	TempC     float64 `json:"temp_c"`
	UsagePct  float64 `json:"usage_pct"`
	MemUsedMB float64 `json:"mem_used_mb"`
	PowerW    float64 `json:"power_w"`
}

type MetricRow struct {
	Time   time.Time         `json:"time"`
	Metric string            `json:"metric"`
	Value  float64           `json:"value"`
	Labels map[string]string `json:"labels,omitempty"`
}

func ToMetricRows(p IngestPayload, ts time.Time) []MetricRow {
	var out []MetricRow

	if p.CPU != nil {
		out = append(out, MetricRow{Time: ts, Metric: "cpu.temp_c", Value: p.CPU.TempC})
		out = append(out, MetricRow{Time: ts, Metric: "cpu.usage_pct", Value: p.CPU.UsagePct})
		out = append(out, MetricRow{Time: ts, Metric: "cpu.load_avg", Value: p.CPU.LoadAvg})
		if p.CPU.FreqMHz > 0 {
			out = append(out, MetricRow{Time: ts, Metric: "cpu.freq_mhz", Value: p.CPU.FreqMHz})
		}
		if p.CPU.Cores > 0 {
			out = append(out, MetricRow{Time: ts, Metric: "cpu.cores", Value: float64(p.CPU.Cores)})
		}
	}

	for _, gpu := range p.GPUs {
		labels := map[string]string{"gpu": gpu.Name}
		out = append(out, MetricRow{Time: ts, Metric: "gpu.temp_c", Value: gpu.TempC, Labels: labels})
		out = append(out, MetricRow{Time: ts, Metric: "gpu.usage_pct", Value: gpu.UsagePct, Labels: labels})
		out = append(out, MetricRow{Time: ts, Metric: "gpu.mem_used_mb", Value: gpu.MemUsedMB, Labels: labels})
		out = append(out, MetricRow{Time: ts, Metric: "gpu.power_w", Value: gpu.PowerW, Labels: labels})
	}

	if len(p.Tags) > 0 {
		for i := range out {
			if out[i].Labels == nil {
				out[i].Labels = map[string]string{}
			}
			for k, v := range p.Tags {
				out[i].Labels[k] = v
			}
		}
	}

	return out
}
