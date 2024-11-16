package main

import (
	"fmt"
	"context"
	_ "net/http/pprof"

	"github.com/ry461ch/metric-collector/internal/app/server"
	config "github.com/ry461ch/metric-collector/internal/config/server"
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
	server := server.New(config.New())
	server.Run(context.Background())
}
