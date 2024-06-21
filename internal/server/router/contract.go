package router

import (
	"net/http"
)

type metricHandlers interface {
	PostPlainGaugeHandler(res http.ResponseWriter, req *http.Request)
	PostPlainCounterHandler(res http.ResponseWriter, req *http.Request)
	GetPlainCounterHandler(res http.ResponseWriter, req *http.Request)
	GetPlainGaugeHandler(res http.ResponseWriter, req *http.Request)
	GetPlainAllMetricsHandler(res http.ResponseWriter, req *http.Request)
	PostJsonHandler(res http.ResponseWriter, req *http.Request)
	GetJsonHandler(res http.ResponseWriter, req *http.Request)
}
