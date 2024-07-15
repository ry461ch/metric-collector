package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/internal/app/server/config"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	"github.com/ry461ch/metric-collector/internal/metricservice"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/models/response"
)

type Handlers struct {
	config  *config.Config
	metricService *metricservice.MetricService
	fileWorker  *fileworker.FileWorker
}

func New(config *config.Config, metricService *metricservice.MetricService, fileWorker *fileworker.FileWorker) *Handlers {
	return &Handlers{
		metricService: metricService,
		config: config,
		fileWorker: fileWorker,
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
	
	DBCtx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
    defer cancel()
	err = h.metricService.SaveMetrics(DBCtx, []metrics.Metrics{metric})
	if err != nil {
		if err.Error() != "INTERNAL_SERVER_ERROR" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if h.config.StoreInterval == int64(0) {
		fileCtx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
    	defer cancel()
		h.fileWorker.ImportToFile(fileCtx)
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
	
	DBCtx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
    defer cancel()
	err = h.metricService.SaveMetrics(DBCtx, []metrics.Metrics{metric})
	if err != nil {
		if err.Error() != "INTERNAL_SERVER_ERROR" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if h.config.StoreInterval == int64(0) {
		fileCtx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
    	defer cancel()
		h.fileWorker.ImportToFile(fileCtx)
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetPlainCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metric := metrics.Metrics{
		ID: metricName,
		MType: "counter",
	}

	DBCtx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
    defer cancel()
	err := h.metricService.GetMetric(DBCtx, &metric)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		if err.Error() == "INVALID_METRIC_TYPE" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
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

	DBCtx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
    defer cancel()
	err := h.metricService.GetMetric(DBCtx, &metric)

	if err != nil {
		if err.Error() == "NOT_FOUND" {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		if err.Error() == "INVALID_METRIC_TYPE" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.WriteString(res, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
}

func (h *Handlers) GetPlainAllMetricsHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	DBCtx, cancel := context.WithTimeout(req.Context(), 4*time.Second)
    defer cancel()
	metricList, err := h.metricService.ExtractMetrics(DBCtx)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

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
		resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Can't read input"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	metric := metrics.Metrics{}
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Bad request format"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	DBCtx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
    defer cancel()
	err = h.metricService.SaveMetrics(DBCtx, []metrics.Metrics{metric})
	if err != nil {
		if err.Error() != "INTERNAL_SERVER_ERROR" {
			resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Bad request format"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
		}
		resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Internal Server Error"})
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(resp)
		return
	}

	if h.config.StoreInterval == int64(0) {
		fileCtx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
    	defer cancel()
		err = h.fileWorker.ImportToFile(fileCtx)
		if err != nil {
			resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Internal server error"})
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(resp)
			return
		}
	}

	resp, _ := json.Marshal(response.ResponseEmptyObject{})
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

func (h *Handlers) GetJSONHandler(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Can't read input"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	metric := metrics.Metrics{}
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Bad request format"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	DBCtx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
    defer cancel()
	err = h.metricService.GetMetric(DBCtx, &metric)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Metric not found"})
			res.WriteHeader(http.StatusNotFound)
			res.Write(resp)
			return
		}
		if err.Error() == "INVALID_METRIC_TYPE" {
			resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Bad metric type"})
			res.WriteHeader(http.StatusBadRequest)
			res.Write(resp)
		}
		resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Internal Server Error"})
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(resp)
		return
	}

	resp, err := json.Marshal(metric)
	if err != nil {
		resp, _ := json.Marshal(response.ResponseErrorObject{Detail: "Internal Server Error"})
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(resp)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

func (h *Handlers) Ping(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
    defer cancel()
    if !h.metricService.Ping(ctx) {
		res.WriteHeader(http.StatusInternalServerError)
		return
    }
	res.WriteHeader(http.StatusOK)
}
