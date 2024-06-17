package hndlservice

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/internal/storage"
)

type HandlerService struct {
	MStorage storage.Storage
}

func (hs *HandlerService) PostGaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseFloat(chi.URLParam(req, "value"), 64)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	hs.MStorage.UpdateGaugeValue(metricName, metricVal)
	res.WriteHeader(http.StatusOK)
}

func (hs *HandlerService) PostCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseInt(chi.URLParam(req, "value"), 10, 0)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	hs.MStorage.UpdateCounterValue(metricName, metricVal)
	res.WriteHeader(http.StatusOK)
}

func (hs *HandlerService) GetCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	val, ok := hs.MStorage.GetCounterValue(metricName)

	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	io.WriteString(res, strconv.FormatInt(val, 10))
}

func (hs *HandlerService) GetGaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	val, ok := hs.MStorage.GetGaugeValue(metricName)

	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	io.WriteString(res, strconv.FormatFloat(val, 'f', -1, 64))
}

func (hs *HandlerService) GetAllMetricsHandler(res http.ResponseWriter, req *http.Request) {
	gaugeMetrics := hs.MStorage.GetGaugeValues()
	counterMetrics := hs.MStorage.GetCounterValues()

	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	for name, val := range gaugeMetrics {
		io.WriteString(res, name+" : "+strconv.FormatFloat(val, 'f', -1, 64)+"\n")
	}
	for name, val := range counterMetrics {
		io.WriteString(res, name+" : "+strconv.FormatInt(val, 10)+"\n")
	}
}
