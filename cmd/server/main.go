package main

import (
	"github.com/ry461ch/metric-collector/internal/app/server"
	config "github.com/ry461ch/metric-collector/internal/config/server"
)

func main() {
	server := server.New(config.New())
	server.Run()
}
