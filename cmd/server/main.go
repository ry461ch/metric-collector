package main

import (
	"flag"
	"net/http"
    "strconv"
	"log"
	"strings"

	"github.com/caarlos0/env/v6"

	"github.com/ry461ch/metric-collector/internal/storage"
	"github.com/ry461ch/metric-collector/internal/net_addr"
)

type Config struct {
	Address		string 		`env:"ADDRESS"`
}


func main() {
	addr := netaddr.NetAddress{Host: "localhost", Port: 8080}
    flag.Var(&addr, "a", "Net address host:port")
    flag.Parse()

	var cfg Config
    err := env.Parse(&cfg)
    if err != nil {
        log.Fatalf("Can't parse env variables: %s", err.Error())
    }
	if cfg.Address != "" {
		addrParts := strings.Split(cfg.Address, ":")
		port, _ := strconv.ParseInt(addrParts[1], 10, 0)
		addr = netaddr.NetAddress{Host: addrParts[0], Port: port}
	}

	server := MetricUpdateServer{mStorage: &storage.MetricStorage{}}
	router := server.MakeRouter()
	err = http.ListenAndServe(addr.Host + ":" + strconv.FormatInt(addr.Port, 10), router)
	if err != nil {
		panic(err)
	}
}
