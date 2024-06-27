package config

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"

	"github.com/ry461ch/metric-collector/internal/config/netaddr"
)

type (
	Config struct {
		Address        string `env:"ADDRESS"`
		ReportInterval string `env:"REPORT_INTERVAL"`
		PollInterval   string `env:"POLL_INTERVAL"`
	}

	Options struct {
		ReportIntervalSec int64
		PollIntervalSec   int64
		Addr              netaddr.NetAddress
	}
)

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
	if cfg.ReportInterval != "" {
		options.ReportIntervalSec, _ = strconv.ParseInt(cfg.ReportInterval, 10, 0)
	}
	if cfg.PollInterval != "" {
		options.PollIntervalSec, _ = strconv.ParseInt(cfg.PollInterval, 10, 0)
	}
}
