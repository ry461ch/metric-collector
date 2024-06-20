package config

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap/zapcore"

	"github.com/ry461ch/metric-collector/internal/config/netaddr"
)

type Config struct {
	Address string `env:"ADDRESS"`
	LogLevelStr string `env:"LOG_LEVEL"`
}

type Options struct {
	Addr netaddr.NetAddress
	LogLevel zapcore.Level
}

func ParseLogLevel(logLevel string) zapcore.Level {
	switch logLevel {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	default:
		return zapcore.InvalidLevel
	}
}

func ParseArgs(opt *Options) {
	flag.Var(&opt.Addr, "a", "Net address host:port")
	logLevelStr := flag.String("l", "INFO", "Log Level")
	flag.Parse()
	opt.LogLevel = ParseLogLevel(*logLevelStr)
}

func ParseEnv(opt *Options) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err.Error())
	}
	if cfg.Address != "" {
		addrParts := strings.Split(cfg.Address, ":")
		port, _ := strconv.ParseInt(addrParts[1], 10, 0)
		opt.Addr.Host = addrParts[0]
		opt.Addr.Port = port
	}
	if cfg.LogLevelStr != "" {
		opt.LogLevel = ParseLogLevel(cfg.LogLevelStr)
	}
}
