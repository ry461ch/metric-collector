package main

import (
	"time"

	"github.com/ry461ch/metric-collector/internal/agent"
	"github.com/ry461ch/metric-collector/internal/models/agent_models"
	"github.com/ry461ch/metric-collector/internal/models/netaddr"
	"github.com/ry461ch/metric-collector/internal/storage/metric_storage"
)

func main() {
	options := agent_models.Options{Addr: netaddr.NetAddress{Host: "localhost", Port: 8080}}
	agent_models.ParseArgs(&options)
	agent_models.ParseEnv(&options)

	mAgent := agent.MetricAgent{
		MStorage:  &metric_storage.MetricStorage{},
		Options:   options,
		TimeState: &agent_models.TimeState{},
	}
	for {
		mAgent.Run()
		time.Sleep(time.Second)
	}
}
