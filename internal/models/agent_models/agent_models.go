package agent_models

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
}

type Options struct {
	ReportIntervalSec int64
	PollIntervalSec   int64
	Addr              netaddr.NetAddress
}

type TimeState struct {
	LastCollectMetricTime time.Time
	LastSendMetricTime    time.Time
}

func ParseArgs(options *Options) {
	flag.Var(&options.Addr, "a", "Net address host:port")
	flag.Int64Var(&options.ReportIntervalSec, "r", 10, "Interval of sending metrics to the server")
	flag.Int64Var(&options.PollIntervalSec, "p", 2, "Interval of polling metrics from runtime")
	flag.Parse()
}

func ParseEnv(options *Options) {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err)
	}
	if cfg.Address != "" {
		addrParts := strings.Split(cfg.Address, ":")
		port, _ := strconv.ParseInt(addrParts[1], 10, 0)
		options.Addr = netaddr.NetAddress{Host: addrParts[0], Port: port}
	}
	if cfg.ReportInterval != 0 {
		options.ReportIntervalSec = cfg.ReportInterval
	}
	if cfg.PollInterval != 0 {
		options.PollIntervalSec = cfg.PollInterval
	}
}
