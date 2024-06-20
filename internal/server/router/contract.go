package router

import (
	"net/http"
)

type metricHandlers interface {
	PostGaugeHandler(res http.ResponseWriter, req *http.Request)
	PostCounterHandler(res http.ResponseWriter, req *http.Request)
	GetCounterHandler(res http.ResponseWriter, req *http.Request)
	GetGaugeHandler(res http.ResponseWriter, req *http.Request)
	GetAllMetricsHandler(res http.ResponseWriter, req *http.Request)
}
