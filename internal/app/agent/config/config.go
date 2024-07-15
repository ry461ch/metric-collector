package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
)

type Config struct {
	ReportIntervalSec int64				`env:"REPORT_INTERVAL"`
	PollIntervalSec   int64				`env:"POLL_INTERVAL"`
	Addr              netaddr.NetAddress	`env:"ADDRESS"`
}

func NewConfig() *Config {
	addr := netaddr.NetAddress{Host: "localhost", Port: 8080}
	return &Config{ReportIntervalSec: 10, PollIntervalSec: 2, Addr: addr}
}


func ParseArgs(cfg *Config) {
	flag.Var(&cfg.Addr, "a", "Net address host:port")
	flag.Int64Var(&cfg.ReportIntervalSec, "r", 10, "Interval of sending metrics to the server")
	flag.Int64Var(&cfg.PollIntervalSec, "p", 2, "Interval of polling metrics from runtime")
	flag.Parse()
}

func ParseEnv(cfg *Config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err)
	}
}
