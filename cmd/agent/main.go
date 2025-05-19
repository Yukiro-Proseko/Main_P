package main

import (
	"paral/internal/agent"
	"paral/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	for i := 0; i < cfg.ComputingPower; i++ {
		go agent.StartWorker()
	}

	select {}
}
