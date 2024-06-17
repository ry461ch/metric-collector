package main

import (
	"net/http"
	"strconv"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
	"github.com/ry461ch/metric-collector/internal/server/parsers"
	"github.com/ry461ch/metric-collector/internal/server/handler_service"
	"github.com/ry461ch/metric-collector/internal/server/router"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

func main() {
	addr := netaddr.NetAddress{Host: "localhost", Port: 8080}
	parsers.ParseArgs(&addr)
	parsers.ParseEnv(&addr)

	handlerService := hndlservice.HandlerService{MStorage: &memstorage.MemStorage{}}
	router := router.Route(&handlerService)
	err := http.ListenAndServe(addr.Host+":"+strconv.FormatInt(addr.Port, 10), router)
	if err != nil {
		panic(err)
	}
}
