package collectors

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"

	"home-telemetry/agent/internal/types"
)

func CollectNvidia() ([]types.GPUMetrics, error) {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=name,temperature.gpu,utilization.gpu,memory.used,power.draw",
		"--format=csv,noheader,nounits",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var gpus []types.GPUMetrics
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) < 5 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		temp := parseFloat(parts[1])
		util := parseFloat(parts[2])
		mem := parseFloat(parts[3])
		power := parseFloat(parts[4])

		gpus = append(gpus, types.GPUMetrics{
			Name:      name,
			TempC:     temp,
			UsagePct:  util,
			MemUsedMB: mem,
			PowerW:    power,
		})
	}

	return gpus, nil
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}