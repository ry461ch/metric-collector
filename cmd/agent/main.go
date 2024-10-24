package main

import (
	_ "net/http/pprof"

	"github.com/ry461ch/metric-collector/internal/app/agent"
	config "github.com/ry461ch/metric-collector/internal/config/agent"
)

func main() {
	agent := agent.New(config.New())
	agent.Run()
}
