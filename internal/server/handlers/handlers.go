package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/internal/metricservice"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/models/response"
	"github.com/ry461ch/metric-collector/internal/server/config"
	"github.com/ry461ch/metric-collector/internal/storage"
	"github.com/ry461ch/metric-collector/internal/fileworker"
)

type Handlers struct {
	options  config.Options
	metricService *metricservice.MetricService
	fileWorker  *fileworker.FileWorker
}

func New(metricStorage storage.Storage, options config.Options) *Handlers {
	metricService := metricservice.New(metricStorage)
	return &Handlers{
		metricService: metricService,
		options: options,
		fileWorker: fileworker.New(options.FileStoragePath, metricStorage),
	}
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
	h.metricService.SaveMetrics([]metrics.Metrics{metric})
	if h.options.StoreInterval == int64(0) {
		h.fileWorker.ImportToFile()
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
	h.metricService.SaveMetrics([]metrics.Metrics{metric})
	if h.options.StoreInterval == int64(0) {
		h.fileWorker.ImportToFile()
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetPlainCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metric := metrics.Metrics{
		ID: metricName,
		MType: "counter",
	}
	err := h.metricService.GetMetric(&metric)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	io.WriteString(res, strconv.FormatInt(*metric.Delta, 10))
}

func (h *Handlers) GetPlainGaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metric := metrics.Metrics{
		ID: metricName,
		MType: "gauge",
	}
	err := h.metricService.GetMetric(&metric)

	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	io.WriteString(res, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
}

func (h *Handlers) GetPlainAllMetricsHandler(res http.ResponseWriter, req *http.Request) {
	metricList := h.metricService.ExtractMetrics()

	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	for _, metric := range metricList {
		switch metric.MType {
		case "counter":
			io.WriteString(res, metric.ID + " : " + strconv.FormatInt(*metric.Delta, 10) + "\n")
		case "gauge":
			io.WriteString(res, metric.ID + " : " + strconv.FormatFloat(*metric.Value, 'f', -1, 64) + "\n")
		default:
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
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

	err = h.metricService.SaveMetrics([]metrics.Metrics{metric})
	if err != nil {
		resp, _ := json.Marshal(response.ErrorObject{Detail: "bad request format"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	if h.options.StoreInterval == int64(0) {
		err = h.fileWorker.ImportToFile()
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

	err = h.metricService.GetMetric(&metric)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			resp, _ := json.Marshal(response.ErrorObject{Detail: "metric not found"})
			res.WriteHeader(http.StatusNotFound)
			res.Write(resp)
			return
		}
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
