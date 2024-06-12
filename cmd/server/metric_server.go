package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ry461ch/metric-collector/internal/storage"
)

type MetricUpdateServer struct {
	m_storage storage.Storage
}

func (server *MetricUpdateServer) gaugeHandler(res http.ResponseWriter, req *http.Request) {
	args := strings.Split(req.URL.Path, "/")
	val, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	server.m_storage.UpdateGaugeValue(args[0], val)
	res.WriteHeader(http.StatusOK)
}

func (server *MetricUpdateServer) counterHandler(res http.ResponseWriter, req *http.Request) {
	args := strings.Split(req.URL.Path, "/")
	val, err := strconv.ParseInt(args[1], 10, 0)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	server.m_storage.UpdateCounterValue(args[0], val)
	res.WriteHeader(http.StatusOK)
}

func (server *MetricUpdateServer) UpdateMetricHandler(res http.ResponseWriter, req *http.Request) {
	args := strings.Split(req.URL.Path, "/")
	if len(args) < 3 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if args[2] == "gauge" {
		http.StripPrefix(`/update/gauge/`, Conveyor(
			http.HandlerFunc(server.gaugeHandler),
			middlewareAllowMethodPost,
			middlewareValidateArgsNum)).ServeHTTP(res, req)
		return
	}
	if args[2] == "counter" {
		http.StripPrefix(`/update/counter/`, Conveyor(
			http.HandlerFunc(server.counterHandler),
			middlewareAllowMethodPost,
			middlewareValidateArgsNum)).ServeHTTP(res, req)
		return
	}
	res.WriteHeader(http.StatusBadRequest)
}