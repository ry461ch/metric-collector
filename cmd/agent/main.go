package main

import (
	config "github.com/ry461ch/metric-collector/internal/config/agent"
	"github.com/ry461ch/metric-collector/internal/app/agent"
)

func main() {
	agent := agent.NewAgent(config.NewConfig())
	agent.Run()
}
