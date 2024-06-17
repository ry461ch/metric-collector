package main

import (
	"net/http"
	"strconv"

	"github.com/ry461ch/metric-collector/internal/models/netaddr"
	"github.com/ry461ch/metric-collector/internal/models/server_models"
	"github.com/ry461ch/metric-collector/internal/server/handler_service"
	"github.com/ry461ch/metric-collector/internal/server/router"
	"github.com/ry461ch/metric-collector/internal/storage/metric_storage"
)

func main() {
	addr := netaddr.NetAddress{Host: "localhost", Port: 8080}
	server_models.ParseArgs(&addr)
	server_models.ParseEnv(&addr)

	handlerService := handler_service.HandlerService{MStorage: &metric_storage.MetricStorage{}}
	router := router.Route(&handlerService)
	err := http.ListenAndServe(addr.Host+":"+strconv.FormatInt(addr.Port, 10), router)
	if err != nil {
		panic(err)
	}
}
