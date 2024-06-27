package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/internal/helpers/metricfilehelper"
	"github.com/ry461ch/metric-collector/internal/helpers/metricmodelshelper"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/models/response"
	"github.com/ry461ch/metric-collector/internal/server/config"
)

type Handlers struct {
	options  config.Options
	mStorage storage
}

func New(mStorage storage, options config.Options) Handlers {
	return Handlers{mStorage: mStorage, options: options}
}

func (h *Handlers) PostPlainGaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseFloat(chi.URLParam(req, "value"), 64)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metric := metrics.Metrics{
		ID:    metricName,
		MType: "gauge",
		Value: &metricVal,
	}
	metricmodelshelper.SaveMetrics([]metrics.Metrics{metric}, h.mStorage)
	if h.options.StoreInterval == int64(0) {
		metricfilehelper.SaveToFile(h.options.FileStoragePath, h.mStorage)
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) PostPlainCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseInt(chi.URLParam(req, "value"), 10, 0)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metric := metrics.Metrics{
		ID:    metricName,
		MType: "counter",
		Delta: &metricVal,
	}
	metricmodelshelper.SaveMetrics([]metrics.Metrics{metric}, h.mStorage)
	if h.options.StoreInterval == int64(0) {
		metricfilehelper.SaveToFile(h.options.FileStoragePath, h.mStorage)
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetPlainCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	val, ok := h.mStorage.GetCounterValue(metricName)

	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	io.WriteString(res, strconv.FormatInt(val, 10))
}

func (h *Handlers) GetPlainGaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	val, ok := h.mStorage.GetGaugeValue(metricName)

	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	io.WriteString(res, strconv.FormatFloat(val, 'f', -1, 64))
}

func (h *Handlers) GetPlainAllMetricsHandler(res http.ResponseWriter, req *http.Request) {
	gaugeMetrics := h.mStorage.GetGaugeValues()
	counterMetrics := h.mStorage.GetCounterValues()

	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	for name, val := range gaugeMetrics {
		io.WriteString(res, name+" : "+strconv.FormatFloat(val, 'f', -1, 64)+"\n")
	}
	for name, val := range counterMetrics {
		io.WriteString(res, name+" : "+strconv.FormatInt(val, 10)+"\n")
	}
}

func (h *Handlers) PostJSONHandler(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		resp, _ := json.Marshal(response.ErrorObject{Detail: "can't read input"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	metric := metrics.Metrics{}
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		resp, _ := json.Marshal(response.ErrorObject{Detail: "bad request format"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	err = metricmodelshelper.SaveMetrics([]metrics.Metrics{metric}, h.mStorage)
	if err != nil {
		resp, _ := json.Marshal(response.ErrorObject{Detail: "bad request format"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	if h.options.StoreInterval == int64(0) {
		err = metricfilehelper.SaveToFile(h.options.FileStoragePath, h.mStorage)
		if err != nil {
			resp, _ := json.Marshal(response.ErrorObject{Detail: "Internal server error"})
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(resp)
			return
		}
	}

	resp, _ := json.Marshal(response.EmptyObject{})
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

func (h *Handlers) GetJSONHandler(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		resp, _ := json.Marshal(response.ErrorObject{Detail: "can't read input"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	metric := metrics.Metrics{}
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		resp, _ := json.Marshal(response.ErrorObject{Detail: "bad request format"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	switch metric.MType {
	case "gauge":
		val, ok := h.mStorage.GetGaugeValue(metric.ID)
		if !ok {
			resp, _ := json.Marshal(response.ErrorObject{Detail: "gauge metric not found"})
			res.WriteHeader(http.StatusNotFound)
			res.Write(resp)
			return
		}
		metric.Value = &val
	case "counter":
		val, ok := h.mStorage.GetCounterValue(metric.ID)
		if !ok {
			resp, _ := json.Marshal(response.ErrorObject{Detail: "counter metric not found"})
			res.WriteHeader(http.StatusNotFound)
			res.Write(resp)
			return
		}
		metric.Delta = &val
	default:
		resp, _ := json.Marshal(response.ErrorObject{Detail: "bad metric type"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	resp, err := json.Marshal(metric)
	if err != nil {
		resp, _ := json.Marshal(response.ErrorObject{Detail: "Internal server error"})
		res.Write(resp)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
