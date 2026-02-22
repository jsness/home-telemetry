package main

import (
	"encoding/json"
	"log"
	"time"

	"home-telemetry/agent/internal/client"
	"home-telemetry/agent/internal/collectors"
	"home-telemetry/agent/internal/config"
	"home-telemetry/agent/internal/types"
)

func main() {
	cfg := config.Load()
	logger := log.New(log.Writer(), "", log.LstdFlags)
	c := client.New(cfg.ServerURL, cfg.AuthToken)

	logger.Printf("agent starting: node=%s interval=%s server=%s", cfg.NodeID, cfg.Interval, cfg.ServerURL)

	collectOnce := func() {
		cpu, cpuErr := collectors.CollectCPU(cfg.LHMURL)
		gpus, gpuErr := collectors.CollectNvidia()

		if cpuErr != nil {
			logger.Printf("cpu error: %v", cpuErr)
		}
		if gpuErr != nil {
			logger.Printf("gpu error: %v", gpuErr)
		}

		payload := types.NewPayload(cfg.NodeID, cpu, gpus)

		if cfg.PrintOnly {
			b, _ := json.MarshalIndent(payload, "", "  ")
			logger.Println(string(b))
			return
		}

		if err := c.Send(payload); err != nil {
			logger.Printf("send error: %v", err)
		}
	}

	if cfg.Once {
		collectOnce()
		return
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		collectOnce()
		<-ticker.C
	}
}