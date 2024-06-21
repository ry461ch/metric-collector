package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/models/response"
)

type Handlers struct {
	mStorage storage
}

func New(mStorage storage) Handlers {
	return Handlers{mStorage: mStorage}
}

func (h *Handlers) PostPlainGaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseFloat(chi.URLParam(req, "value"), 64)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	h.mStorage.UpdateGaugeValue(metricName, metricVal)
	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) PostPlainCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseInt(chi.URLParam(req, "value"), 10, 0)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	h.mStorage.UpdateCounterValue(metricName, metricVal)
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

	if metric.ID == "" || metric.MType == "" {
		resp, _ := json.Marshal(response.ErrorObject{Detail: "empty name or type"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}

	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			resp, _ := json.Marshal(response.ErrorObject{Detail: "empty value for gauge"})
			res.WriteHeader(http.StatusBadRequest)
			res.Write(resp)
			return
		}
		h.mStorage.UpdateGaugeValue(metric.ID, *metric.Value)
	case "counter":
		if metric.Delta == nil {
			resp, _ := json.Marshal(response.ErrorObject{Detail: "empty value for counter"})
			res.WriteHeader(http.StatusBadRequest)
			res.Write(resp)
			return
		}
		h.mStorage.UpdateCounterValue(metric.ID, *metric.Delta)
	default:
		resp, _ := json.Marshal(response.ErrorObject{Detail: "unexpected metric type"})
		res.WriteHeader(http.StatusBadRequest)
		res.Write(resp)
		return
	}
	resp, err := json.Marshal(response.EmptyObject{})
	if err != nil {
		resp, _ := json.Marshal(response.ErrorObject{Detail: "Internal server error"})
        res.WriteHeader(http.StatusInternalServerError)
		res.Write(resp)
        return
    }
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
