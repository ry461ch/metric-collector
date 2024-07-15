package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"

	"github.com/ry461ch/metric-collector/internal/app/server/config"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	"github.com/ry461ch/metric-collector/internal/metricservice"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

func mockRouter(handlers *Handlers) chi.Router {
	router := chi.NewRouter()
	router.Post("/update/counter/{name}/{value}", handlers.PostPlainCounterHandler)
	router.Post("/update/gauge/{name}/{value}", handlers.PostPlainGaugeHandler)
	router.Post("/update/", handlers.PostJSONHandler)
	router.Get("/value/counter/{name}", handlers.GetPlainCounterHandler)
	router.Get("/value/gauge/{name}", handlers.GetPlainGaugeHandler)
	router.Post("/value/", handlers.GetJSONHandler)
	router.Get("/", handlers.GetPlainAllMetricsHandler)
	return router
}

func TestPostTextGaugeHandler(t *testing.T) {
	memStorage := memstorage.MemStorage{}

	metricService := metricservice.New(&memStorage)
	fileWorker := fileworker.New("", metricService)
	handlers := New(&config.Config{StoreInterval: 1}, metricService, fileWorker, nil)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	_, err := client.R().Post(srv.URL + "/update/gauge/some_metric/10.0")
	assert.Nil(t, err, "Сервер вернул 500")

	_, err = client.R().Post(srv.URL + "/update/gauge/some_metric/12.0")
	assert.Nil(t, err, "Сервер вернул 500")

	val, _ := memStorage.GetGaugeValue("some_metric")
	assert.Equal(t, float64(12.0), val, "Сохраненное значение метрики типа gauge не совпадает с ожидаемым")
}

func TestPostTextCounterHandler(t *testing.T) {
	memStorage := memstorage.MemStorage{}

	metricService := metricservice.New(&memStorage)
	fileWorker := fileworker.New("", metricService)
	handlers := New(&config.Config{StoreInterval: 1}, metricService, fileWorker, nil)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	_, err := client.R().Post(srv.URL + "/update/counter/some_metric/10")
	assert.Nil(t, err, "Сервер вернул 500")

	_, err = client.R().Post(srv.URL + "/update/counter/some_metric/12")
	assert.Nil(t, err, "Сервер вернул 500")

	val, _ := memStorage.GetCounterValue("some_metric")
	assert.Equal(t, int64(22), val, "Сохраненное значение метрики типа counter не совпадает с ожидаемым")
}

func TestGetTextGaugeHandler(t *testing.T) {
	memStorage := memstorage.MemStorage{}
	memStorage.UpdateGaugeValue("some_metric", 10.5)

	metricService := metricservice.New(&memStorage)
	fileWorker := fileworker.New("", metricService)
	handlers := New(&config.Config{StoreInterval: 1}, metricService, fileWorker, nil)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, err := client.R().Get(srv.URL + "/value/gauge/some_metric")
	assert.Nil(t, err, "Сервер вернул 500")

	body := resp.Body()
	assert.Equal(t, "10.5", string(body), "Неверное значение метрики gauge")

	resp, err = client.R().Get(srv.URL + "/value/counter/undefined")
	assert.Nil(t, err, "Сервер вернул 500")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "Нашлась несуществующая метрика")
}

func TestGetTextCounterHandler(t *testing.T) {
	memStorage := memstorage.MemStorage{}
	memStorage.UpdateCounterValue("some_metric", 10)

	metricService := metricservice.New(&memStorage)
	fileWorker := fileworker.New("", metricService)
	handlers := New(&config.Config{StoreInterval: 1}, metricService, fileWorker, nil)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, err := client.R().Get(srv.URL + "/value/counter/some_metric")
	assert.Nil(t, err, "Сервер вернул 500")

	body := resp.Body()
	assert.Equal(t, "10", string(body), "Неверное значение метрики counter")

	resp, err = client.R().Get(srv.URL + "/value/counter/undefined")
	assert.Nil(t, err, "Сервер вернул 500")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "Нашлась несуществующая метрика")
}

func TestGetAllMetricsHandler(t *testing.T) {
	memStorage := memstorage.MemStorage{}
	memStorage.UpdateCounterValue("counter_1", 1)
	memStorage.UpdateCounterValue("counter_2", 2)
	memStorage.UpdateGaugeValue("gauge_1", 1)
	memStorage.UpdateGaugeValue("gauge_2", 2)

	expectedBody := "counter_1 : 1\ncounter_2 : 2\ngauge_1 : 1\ngauge_2 : 2\n"

	metricService := metricservice.New(&memStorage)
	fileWorker := fileworker.New("", metricService)
	handlers := New(&config.Config{StoreInterval: 1}, metricService, fileWorker, nil)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, err := client.R().Get(srv.URL + "/")
	assert.Nil(t, err, "Сервер вернул 500")

	header := resp.Header().Get("Content-Type")
	assert.Equal(t, "text/html; charset=utf-8", header, "Неверное значение content-type")

	body := resp.Body()
	assert.Equal(t, len(expectedBody), len(string(body)), "Неверное значение тела ответа")
}

func TestPostJSONHandler(t *testing.T) {
	memStorage := memstorage.MemStorage{}

	metricService := metricservice.New(&memStorage)
	fileWorker := fileworker.New("", metricService)
	handlers := New(&config.Config{StoreInterval: 1}, metricService, fileWorker, nil)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()

	defaultValue := float64(5.5)
	defaultDelta := int64(7)
	testCases := []struct {
		testName     string
		method       string
		requestPath  string
		requestBody  *metrics.Metrics
		expectedCode int
	}{
		{
			testName:    "ok post gauge",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metrics{
				ID:    "test",
				MType: "gauge",
				Value: &defaultValue,
			},
			expectedCode: http.StatusOK,
		},
		{
			testName:    "ok post counter",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metrics{
				ID:    "test",
				MType: "counter",
				Delta: &defaultDelta,
			},
			expectedCode: http.StatusOK,
		},
		{
			testName:    "ok get gauge",
			method:      http.MethodPost,
			requestPath: "/value/",
			requestBody: &metrics.Metrics{
				ID:    "test",
				MType: "gauge",
			},
			expectedCode: http.StatusOK,
		},
		{
			testName:    "ok get counter",
			method:      http.MethodPost,
			requestPath: "/value/",
			requestBody: &metrics.Metrics{
				ID:    "test",
				MType: "counter",
			},
			expectedCode: http.StatusOK,
		},
		{
			testName:    "invalid type for post",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metrics{
				ID:    "test",
				MType: "invalid",
				Value: &defaultValue,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			testName:    "invalid type for get",
			method:      http.MethodPost,
			requestPath: "/value/",
			requestBody: &metrics.Metrics{
				ID:    "test",
				MType: "invalid",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			testName:    "bad value for post gauge",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metrics{
				ID:    "test",
				MType: "gauge",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			testName:    "bad value for post counter",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metrics{
				ID:    "test",
				MType: "counter",
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			req, _ := json.Marshal(tc.requestBody)
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(req).
				Execute(tc.method, srv.URL+tc.requestPath)
			assert.Nil(t, err, "Сервер вернул 500")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestJsonGaugeStorageHandler(t *testing.T) {
	memStorage := memstorage.MemStorage{}

	metricService := metricservice.New(&memStorage)
	fileWorker := fileworker.New("", metricService)
	handlers := New(&config.Config{StoreInterval: 1}, metricService, fileWorker, nil)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()

	val := float64(5.0)
	metric := &metrics.Metrics{
		ID:    "test",
		MType: "gauge",
		Value: &val,
	}

	req, _ := json.Marshal(metric)
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Execute(http.MethodPost, srv.URL+"/update/")
	assert.Nil(t, err, "Сервер вернул 500")
	assert.Equal(t, http.StatusOK, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

	metric.Value = nil
	req, _ = json.Marshal(metric)
	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Execute(http.MethodPost, srv.URL+"/value/")
	assert.Nil(t, err, "Сервер вернул 500")
	assert.Equal(t, http.StatusOK, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

	json.Unmarshal(resp.Body(), metric)
	assert.Equal(t, val, *metric.Value, "Неверно сохранилась метрика")
}
