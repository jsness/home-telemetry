package collectors

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"

	"home-telemetry/agent/internal/types"
)

func CollectCPU(lhmURL string) (*types.CPUMetrics, error) {
	percent, err := cpu.Percent(1*time.Second, false)
	if err != nil || len(percent) == 0 {
		return nil, err
	}

	cores := runtime.NumCPU()
	metrics := &types.CPUMetrics{
		UsagePct: percent[0],
		Cores:    cores,
	}

	if lhmURL != "" {
		if temp, err := CollectCPUTempFromLHM(lhmURL); err == nil {
			metrics.TempC = temp
		}
	}

	return metrics, nil
}