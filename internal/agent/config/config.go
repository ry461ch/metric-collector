package config

import (
	"time"

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
