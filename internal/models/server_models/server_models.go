package server_models

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

func ParseArgs(addr *netaddr.NetAddress) {
	flag.Var(addr, "a", "Net address host:port")
	flag.Parse()
}

func ParseEnv(addr *netaddr.NetAddress) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err.Error())
	}
	if cfg.Address != "" {
		addrParts := strings.Split(cfg.Address, ":")
		port, _ := strconv.ParseInt(addrParts[1], 10, 0)
		addr.Host = addrParts[0]
		addr.Port = port
	}
}
