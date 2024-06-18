package main

import (
	"time"

	"github.com/ry461ch/metric-collector/internal/agent/runner"
	"github.com/ry461ch/metric-collector/internal/agent/parsers"
	"github.com/ry461ch/metric-collector/internal/agent/config"
	"github.com/ry461ch/metric-collector/internal/models/netaddr"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

func main() {
	options := config.Options{Addr: netaddr.NetAddress{Host: "localhost", Port: 8080}}
	parsers.ParseArgs(&options)
	parsers.ParseEnv(&options)

	mAgent := agent.NewAgent(&config.TimeState{}, options, &memstorage.MemStorage{})
	for {
		mAgent.Run()
		time.Sleep(time.Second)
	}
}
