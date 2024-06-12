package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/storage"
)

func TestServer(t *testing.T) {
	default_gauge_request := "/update/gauge/some_metric/10.0"
	default_counter_request := "/update/counter/some_metric/10"

	testCases := []struct {
		test_name    string
		method       string
		request_url  string
		expectedCode int
	}{
		{test_name: "invalid method for gauge", method: http.MethodGet, request_url: default_gauge_request, expectedCode: http.StatusMethodNotAllowed},
		{test_name: "invalid method for counter", method: http.MethodDelete, request_url: default_counter_request, expectedCode: http.StatusMethodNotAllowed},
		{test_name: "ok for gauge", method: http.MethodPost, request_url: default_gauge_request, expectedCode: http.StatusOK},
		{test_name: "ok for counter", method: http.MethodPost, request_url: default_counter_request, expectedCode: http.StatusOK},
		{test_name: "ok with / at the end", method: http.MethodPost, request_url: default_counter_request + "/", expectedCode: http.StatusOK},
		{test_name: "no metric name and value for counter", method: http.MethodPost, request_url: "/update/counter/", expectedCode: http.StatusNotFound},
		{test_name: "no metric name and value for gauge", method: http.MethodPost, request_url: "/update/gauge/", expectedCode: http.StatusNotFound},
		{test_name: "invalid metric type", method: http.MethodPost, request_url: "/update/invalid_metric_type/", expectedCode: http.StatusBadRequest},
		{test_name: "no metric name for counter", method: http.MethodPost, request_url: "/update/counter/10", expectedCode: http.StatusNotFound},
		{test_name: "no metric name for gauge ", method: http.MethodPost, request_url: "/update/gauge/10", expectedCode: http.StatusNotFound},
		{test_name: "invalid metric value for counter", method: http.MethodPost, request_url: "/update/counter/test/10.0", expectedCode: http.StatusBadRequest},
		{test_name: "invalid metric value for gauge", method: http.MethodPost, request_url: "/update/gauge/test/str", expectedCode: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.test_name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.request_url, nil)
			resp := httptest.NewRecorder()

			update_metric_server := MetricUpdateServer{m_storage: &storage.MetricStorage{}}
			update_metric_server.UpdateMetricHandler(resp, req)

			assert.Equal(t, tc.expectedCode, resp.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestGaugeServe(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/some_metric/10.0", nil)
	resp := httptest.NewRecorder()

	storage := storage.MetricStorage{}
	update_metric_server := MetricUpdateServer{m_storage: &storage}
	update_metric_server.UpdateMetricHandler(resp, req)

	// update same metric
	req = httptest.NewRequest(http.MethodPost, "/update/gauge/some_metric/12.0", nil)
	update_metric_server.UpdateMetricHandler(resp, req)

	assert.Equal(t, float64(12.0), storage.GetGaugeValue("some_metric"), "Сохраненное значение метрики типа gauge не совпадает с ожидаемым")
}

func TestCounterServe(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update/counter/some_metric/10", nil)
	resp := httptest.NewRecorder()

	storage := storage.MetricStorage{}
	update_metric_server := MetricUpdateServer{m_storage: &storage}
	update_metric_server.UpdateMetricHandler(resp, req)

	// update same metric
	req = httptest.NewRequest(http.MethodPost, "/update/counter/some_metric/12", nil)
	update_metric_server.UpdateMetricHandler(resp, req)

	assert.Equal(t, int64(22), storage.GetCounterValue("some_metric"), "Сохраненное значение метрики типа counter не совпадает с ожидаемым")
}
