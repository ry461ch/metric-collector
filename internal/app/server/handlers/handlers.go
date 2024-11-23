// Module for handling all metric requests
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"

	config "github.com/ry461ch/metric-collector/internal/config/server"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
)

// @Title Metric API
// @Description Сервис хранения метрик.
// @Version 1.0

// @SecurityDefinitions.apikey SecurityKeyAuth
// @In header
// @Name HashSHA256

// Типы для работы хэндлеров сервера метрик
type (
	// Handlers - обработчики запросов на сохранение/получение метрик
	Handlers struct {
		config        *config.Config
		metricStorage Storage
		fileWorker    FileWorker
	}

	// ResponseEmptyObject - пустой объект для возврата из функций с content-type=application/json
	ResponseEmptyObject struct{}

	// ResponseErrorObject - объект-ошибка для возврата из функций с content-type=application/json
	ResponseErrorObject struct {
		Detail string `json:"detail,omitempty"`
	}
)

// Init metric handlers
func New(config *config.Config, metricStorage Storage, fileWorker FileWorker) *Handlers {
	return &Handlers{
		metricStorage: metricStorage,
		config:        config,
		fileWorker:    fileWorker,
	}
}

func (h *Handlers) saveMetrics(ctx context.Context, metricList []metrics.Metric) error {
	for i := 0; i <= 3; i += 1 {
		DBCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		err := h.metricStorage.SaveMetrics(DBCtx, metricList)
		if err == nil {
			return nil
		}
		if pgerrcode.IsConnectionException(err.Error()) && i != 3 {
			cancel()
			time.Sleep(time.Second * time.Duration(1))
			continue
		}
		if err.Error() == "INVALID_METRIC" {
			return err
		}
		return errors.New("INTERNAL_SERVER_ERROR")
	}
	return nil
}

func (h *Handlers) getMetric(ctx context.Context, metric *metrics.Metric) error {
	for i := 0; i <= 3; i += 1 {
		DBCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		err := h.metricStorage.GetMetric(DBCtx, metric)
		if err == nil {
			return nil
		}
		if pgerrcode.IsConnectionException(err.Error()) && i != 3 {
			cancel()
			time.Sleep(time.Second * time.Duration(1))
			continue
		}
		if err.Error() == "NOT_FOUND" || err.Error() == "INVALID_METRIC_TYPE" {
			return err
		}
		return errors.New("INTERNAL_SERVER_ERROR")
	}
	return nil
}

func (h *Handlers) extractMetrics(ctx context.Context) ([]metrics.Metric, error) {
	for i := 0; i <= 3; i += 1 {
		DBCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
		metricList, err := h.metricStorage.ExtractMetrics(DBCtx)
		if err == nil {
			return metricList, nil
		}
		if pgerrcode.IsConnectionException(err.Error()) && i != 3 {
			cancel()
			time.Sleep(time.Second * time.Duration(1))
			continue
		}
		return nil, errors.New("INTERNAL_SERVER_ERROR")
	}
	return []metrics.Metric{}, nil
}

// PostPlainGaugeHandler godoc
// @Summary Save one metric with gauge type
// @Description Save gauge metric
// @ID storagePostPlainGauge
// @Accept  text/plain
// @Produce text/plain
// @Param name path string true "Metric name"
// @Param value path float64 true "Metric value"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Error"
// @Security SecurityKeyAuth
// @Router /update/gauge/{name}/{value} [post]
func (h *Handlers) PostPlainGaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseFloat(chi.URLParam(req, "value"), 64)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metricList := []metrics.Metric{
		{
			ID:    metricName,
			MType: "gauge",
			Value: &metricVal,
		},
	}

	err = h.saveMetrics(req.Context(), metricList)
	if err != nil {
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

// PostPlainCounterHandler godoc
// @Summary Save one metric with counter type
// @Description Save counter metric
// @ID storagePostPlainCounter
// @Accept  text/plain
// @Produce text/plain
// @Param name path string true "Metric name"
// @Param value path float64 true "Metric value"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Error"
// @Security SecurityKeyAuth
// @Router /update/counter/{name}/{value} [post]
func (h *Handlers) PostPlainCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseInt(chi.URLParam(req, "value"), 10, 0)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metricList := []metrics.Metric{
		{
			ID:    metricName,
			MType: "counter",
			Delta: &metricVal,
		},
	}

	err = h.saveMetrics(req.Context(), metricList)
	if err != nil {
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

// GetPlainCounterHandler godoc
// @Summary Get one metric with counter type
// @Description Get counter metric
// @ID storageGetPlainCounter
// @Accept  text/plain
// @Produce text/plain
// @Param name path string true "Metric name"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Error"
// @Security SecurityKeyAuth
// @Router /value/counter/{name} [get]
func (h *Handlers) GetPlainCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metric := metrics.Metric{
		ID:    metricName,
		MType: "counter",
	}

	err := h.getMetric(req.Context(), &metric)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(res, strconv.FormatInt(*metric.Delta, 10))
}

// GetPlainGaugeHandler godoc
// @Summary Get one metric with gauge type
// @Description Get gauge metric
// @ID storageGetPlainGauge
// @Accept  text/plain
// @Produce text/plain
// @Param name path string true "Metric name"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Error"
// @Security SecurityKeyAuth
// @Router /value/gauge/{name} [get]
func (h *Handlers) GetPlainGaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metric := metrics.Metric{
		ID:    metricName,
		MType: "gauge",
	}

	err := h.getMetric(req.Context(), &metric)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(res, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
}

// GetPlainAllMetricsHandler godoc
// @Summary Get all metrics
// @Description Get all metrics
// @ID storageGetPlainAll
// @Accept  text/plain
// @Produce text/plain
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Internal Error"
// @Security SecurityKeyAuth
// @Router / [get]
func (h *Handlers) GetPlainAllMetricsHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	metricList, err := h.extractMetrics(req.Context())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, metric := range metricList {
		switch metric.MType {
		case "counter":
			io.WriteString(res, metric.ID+" : "+strconv.FormatInt(*metric.Delta, 10)+"\n")
		case "gauge":
			io.WriteString(res, metric.ID+" : "+strconv.FormatFloat(*metric.Value, 'f', -1, 64)+"\n")
		default:
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// PostJSONHandler godoc
// @Summary Post json metric
// @Description Post json metric
// @ID storagePostJSONSingle
// @Accept  application/json
// @Produce application/json
// @Param metric body metrics.Metric true "Metric data"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Error"
// @Security SecurityKeyAuth
// @Router /update [post]
func (h *Handlers) PostJSONHandler(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		resp, _ := json.Marshal(ResponseErrorObject{Detail: "Can't read input"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	metric := metrics.Metric{}
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		resp, _ := json.Marshal(ResponseErrorObject{Detail: "Bad request format"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	err = h.saveMetrics(req.Context(), []metrics.Metric{metric})
	if err != nil {
		if err.Error() == "INVALID_METRIC" {
			resp, _ := json.Marshal(ResponseErrorObject{Detail: "Bad request format"})
			res.WriteHeader(http.StatusBadRequest)
			res.Write(resp)
			return
		}
		resp, _ := json.Marshal(ResponseErrorObject{Detail: "Internal Server Error"})
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(resp)
		return
	}

	if h.config.StoreInterval == int64(0) {
		fileCtx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		defer cancel()
		h.fileWorker.ImportToFile(fileCtx)
	}

	resp, _ := json.Marshal(ResponseEmptyObject{})
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

// GetJSONHandler godoc
// @Summary Get json metric
// @Description Get json metric
// @ID storageGetJSON
// @Accept  application/json
// @Produce application/json
// @Param metric body metrics.Metric true "Metric data"
// @Success 200 {object} metrics.Metric
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Error"
// @Security SecurityKeyAuth
// @Router /value [get]
func (h *Handlers) GetJSONHandler(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		resp, _ := json.Marshal(ResponseErrorObject{Detail: "Can't read input"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	metric := metrics.Metric{}
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		resp, _ := json.Marshal(ResponseErrorObject{Detail: "Bad request format"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	err = h.getMetric(req.Context(), &metric)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			resp, _ := json.Marshal(ResponseErrorObject{Detail: "Metric not found"})
			res.WriteHeader(http.StatusNotFound)
			res.Write(resp)
			return
		}
		if err.Error() == "INVALID_METRIC_TYPE" {
			resp, _ := json.Marshal(ResponseErrorObject{Detail: "Bad metric type"})
			res.WriteHeader(http.StatusBadRequest)
			res.Write(resp)
		}
		resp, _ := json.Marshal(ResponseErrorObject{Detail: "Internal Server Error"})
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(resp)
		return
	}

	resp, err := json.Marshal(metric)
	if err != nil {
		respErr, _ := json.Marshal(ResponseErrorObject{Detail: "Internal Server Error"})
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(respErr)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

// Ping godoc
// @Summary Ping server
// @ID infoHealth
// @Accept  json
// @Produce json
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Internal error"
// @Router /ping [get]
func (h *Handlers) Ping(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
	defer cancel()
	if externalStorage, ok := h.metricStorage.(ExternalStorage); ok && !externalStorage.Ping(ctx) {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

// PostMetricsHandler godoc
// @Summary Post json metric s
// @Description Post json metrics
// @ID storagePostJSONSeveral
// @Accept  application/json
// @Produce application/json
// @Param metrics body []metrics.Metric true "Metric data"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Error"
// @Security SecurityKeyAuth
// @Router /updates [post]
func (h *Handlers) PostMetricsHandler(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		resp, _ := json.Marshal(ResponseErrorObject{Detail: "Can't read input"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	metricList := []metrics.Metric{}
	data := buf.Bytes()
	if len(data) == 0 {
		resp, _ := json.Marshal(ResponseErrorObject{Detail: "Empty input"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	err = json.Unmarshal(data, &metricList)
	if err != nil {
		resp, _ := json.Marshal(ResponseErrorObject{Detail: "Can't read input"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	err = h.saveMetrics(req.Context(), metricList)
	if err != nil {
		if err.Error() == "INVALID_METRIC" {
			resp, _ := json.Marshal(ResponseErrorObject{Detail: "Bad request format"})
			res.WriteHeader(http.StatusBadRequest)
			res.Write(resp)
			return
		}
		resp, _ := json.Marshal(ResponseErrorObject{Detail: "Internal Server Error"})
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(resp)
		return
	}

	resp, _ := json.Marshal(ResponseEmptyObject{})
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
