package main

import (
	"flag"
	"net/http"
    "strconv"

	"github.com/ry461ch/metric-collector/internal/storage"
	"github.com/ry461ch/metric-collector/internal/net_addr"
)

func main() {
	addr := new(netaddr.NetAddress)
    _ = flag.Value(addr)
    flag.Var(addr, "a", "Net address host:port")
    flag.Parse()

	server := MetricUpdateServer{mStorage: &storage.MetricStorage{}}
	router := server.MakeRouter()
	err := http.ListenAndServe(addr.Host + ":" + strconv.FormatInt(addr.Port, 10), router)
	if err != nil {
		panic(err)
	}
}
