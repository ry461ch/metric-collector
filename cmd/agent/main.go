package main

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/ry461ch/metric-collector/internal/app/agent"
	config "github.com/ry461ch/metric-collector/internal/config/agent"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
	agent := agent.New(config.New())
	agent.Run()
}
