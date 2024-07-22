package main

import (
	config "github.com/ry461ch/metric-collector/internal/config/server"
	"github.com/ry461ch/metric-collector/internal/app/server"
)

func main() {
	server := server.NewServer(config.NewConfig())
	server.Run()
}
