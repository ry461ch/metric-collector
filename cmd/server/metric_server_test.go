package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"

	"github.com/ry461ch/metric-collector/internal/storage"
)

func TestServer(t *testing.T) {
	defaultGaugeRequest := "/update/gauge/some_metric/10.0"
	defaultCounterRequest := "/update/counter/some_metric/10"

	testCases := []struct {
		testName    string
		method       string
		requestPath  string
		expectedCode int
	}{
		{testName: "invalid method for gauge", method: http.MethodGet, requestPath: defaultGaugeRequest, expectedCode: http.StatusMethodNotAllowed},
		{testName: "invalid method for counter", method: http.MethodDelete, requestPath: defaultCounterRequest, expectedCode: http.StatusMethodNotAllowed},
		{testName: "ok for gauge", method: http.MethodPost, requestPath: defaultGaugeRequest, expectedCode: http.StatusOK},
		{testName: "ok for counter", method: http.MethodPost, requestPath: defaultCounterRequest, expectedCode: http.StatusOK},
		{testName: "no metric name and value for counter", method: http.MethodPost, requestPath: "/update/counter/", expectedCode: http.StatusNotFound},
		{testName: "no metric name and value for gauge", method: http.MethodPost, requestPath: "/update/gauge/", expectedCode: http.StatusNotFound},
		{testName: "invalid metric type", method: http.MethodPost, requestPath: "/update/invalid_metric_type", expectedCode: http.StatusBadRequest},
		{testName: "no metric name for counter", method: http.MethodPost, requestPath: "/update/counter/10", expectedCode: http.StatusNotFound},
		{testName: "no metric name for gauge ", method: http.MethodPost, requestPath: "/update/gauge/10", expectedCode: http.StatusNotFound},
		{testName: "invalid metric value for counter", method: http.MethodPost, requestPath: "/update/counter/test/10.0", expectedCode: http.StatusBadRequest},
		{testName: "invalid metric value for gauge", method: http.MethodPost, requestPath: "/update/gauge/test/str", expectedCode: http.StatusBadRequest},
	}

	client := resty.New()

	updateMetricServer := MetricUpdateServer{mStorage: &storage.MetricStorage{}}
	router := updateMetricServer.MakeRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			resp, err := client.R().Execute(tc.method, srv.URL + tc.requestPath)
			assert.Nil(t, err, "Сервер вернул 500")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestGaugeServe(t *testing.T) {
	memStorage := storage.MetricStorage{}

	updateMetricServer := MetricUpdateServer{mStorage: &memStorage}
	router := updateMetricServer.MakeRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	_, err := client.R().Post(srv.URL + "/update/gauge/some_metric/10.0")
	assert.Nil(t, err, "Сервер вернул 500")

	_, err = client.R().Post(srv.URL + "/update/gauge/some_metric/12.0")
	assert.Nil(t, err, "Сервер вернул 500")

	assert.Equal(t, float64(12.0), memStorage.GetGaugeValue("some_metric"), "Сохраненное значение метрики типа gauge не совпадает с ожидаемым")
}

func TestCounterServe(t *testing.T) {
	memStorage := storage.MetricStorage{}

	updateMetricServer := MetricUpdateServer{mStorage: &memStorage}
	router := updateMetricServer.MakeRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	_, err := client.R().Post(srv.URL + "/update/counter/some_metric/10")
	assert.Nil(t, err, "Сервер вернул 500")

	_, err = client.R().Post(srv.URL + "/update/counter/some_metric/12")
	assert.Nil(t, err, "Сервер вернул 500")

	assert.Equal(t, int64(22), memStorage.GetCounterValue("some_metric"), "Сохраненное значение метрики типа counter не совпадает с ожидаемым")
}
