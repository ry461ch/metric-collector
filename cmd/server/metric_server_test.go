package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/storage"
)

func TestServer(t *testing.T) {
	defaultGaugeRequest := "/update/gauge/some_metric/10.0"
	defaultCounterRequest := "/update/counter/some_metric/10"

	testCases := []struct {
		testName    string
		method       string
		requestUrl  string
		expectedCode int
	}{
		{testName: "invalid method for gauge", method: http.MethodGet, requestUrl: defaultGaugeRequest, expectedCode: http.StatusMethodNotAllowed},
		{testName: "invalid method for counter", method: http.MethodDelete, requestUrl: defaultCounterRequest, expectedCode: http.StatusMethodNotAllowed},
		{testName: "ok for gauge", method: http.MethodPost, requestUrl: defaultGaugeRequest, expectedCode: http.StatusOK},
		{testName: "ok for counter", method: http.MethodPost, requestUrl: defaultCounterRequest, expectedCode: http.StatusOK},
		{testName: "ok with / at the end", method: http.MethodPost, requestUrl: defaultCounterRequest + "/", expectedCode: http.StatusOK},
		{testName: "no metric name and value for counter", method: http.MethodPost, requestUrl: "/update/counter/", expectedCode: http.StatusNotFound},
		{testName: "no metric name and value for gauge", method: http.MethodPost, requestUrl: "/update/gauge/", expectedCode: http.StatusNotFound},
		{testName: "invalid metric type", method: http.MethodPost, requestUrl: "/update/invalid_metric_type/", expectedCode: http.StatusBadRequest},
		{testName: "no metric name for counter", method: http.MethodPost, requestUrl: "/update/counter/10", expectedCode: http.StatusNotFound},
		{testName: "no metric name for gauge ", method: http.MethodPost, requestUrl: "/update/gauge/10", expectedCode: http.StatusNotFound},
		{testName: "invalid metric value for counter", method: http.MethodPost, requestUrl: "/update/counter/test/10.0", expectedCode: http.StatusBadRequest},
		{testName: "invalid metric value for gauge", method: http.MethodPost, requestUrl: "/update/gauge/test/str", expectedCode: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.requestUrl, nil)
			resp := httptest.NewRecorder()

			updateMetricServer := MetricUpdateServer{mStorage: &storage.MetricStorage{}}
			updateMetricServer.UpdateMetricHandler(resp, req)

			assert.Equal(t, tc.expectedCode, resp.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestGaugeServe(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/some_metric/10.0", nil)
	resp := httptest.NewRecorder()

	storage := storage.MetricStorage{}
	updateMetricServer := MetricUpdateServer{mStorage: &storage}
	updateMetricServer.UpdateMetricHandler(resp, req)

	// update same metric
	req = httptest.NewRequest(http.MethodPost, "/update/gauge/some_metric/12.0", nil)
	updateMetricServer.UpdateMetricHandler(resp, req)

	assert.Equal(t, float64(12.0), storage.GetGaugeValue("some_metric"), "Сохраненное значение метрики типа gauge не совпадает с ожидаемым")
}

func TestCounterServe(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update/counter/some_metric/10", nil)
	resp := httptest.NewRecorder()

	storage := storage.MetricStorage{}
	updateMetricServer := MetricUpdateServer{mStorage: &storage}
	updateMetricServer.UpdateMetricHandler(resp, req)

	// update same metric
	req = httptest.NewRequest(http.MethodPost, "/update/counter/some_metric/12", nil)
	updateMetricServer.UpdateMetricHandler(resp, req)

	assert.Equal(t, int64(22), storage.GetCounterValue("some_metric"), "Сохраненное значение метрики типа counter не совпадает с ожидаемым")
}
