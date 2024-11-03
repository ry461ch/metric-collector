package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"

	config "github.com/ry461ch/metric-collector/internal/config/server"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	memstorage "github.com/ry461ch/metric-collector/internal/storage/memory"
)

func mockRouter(handlers *Handlers) chi.Router {
	router := chi.NewRouter()
	router.Post("/update/counter/{name}/{value}", handlers.PostPlainCounterHandler)
	router.Post("/update/gauge/{name}/{value}", handlers.PostPlainGaugeHandler)
	router.Post("/update/", handlers.PostJSONHandler)
	router.Post("/updates/", handlers.PostMetricsHandler)
	router.Get("/value/counter/{name}", handlers.GetPlainCounterHandler)
	router.Get("/value/gauge/{name}", handlers.GetPlainGaugeHandler)
	router.Post("/value/", handlers.GetJSONHandler)
	router.Get("/", handlers.GetPlainAllMetricsHandler)
	return router
}

type InvalidStorage struct{}

func (is *InvalidStorage) ExtractMetrics(ctx context.Context) ([]metrics.Metric, error) {
	return nil, errors.New("INTERNAL_SERVER_ERROR")
}
func (is *InvalidStorage) SaveMetrics(ctx context.Context, metricList []metrics.Metric) error {
	return errors.New("INTERNAL_SERVER_ERROR")
}
func (is *InvalidStorage) GetMetric(ctx context.Context, metric *metrics.Metric) error {
	return errors.New("INTERNAL_SERVER_ERROR")
}

func TestPostTextGaugeHandler(t *testing.T) {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	_, err := client.R().Post(srv.URL + "/update/gauge/some_metric/10.0")
	assert.Nil(t, err, "Сервер вернул 500")

	_, err = client.R().Post(srv.URL + "/update/gauge/some_metric/12.0")
	assert.Nil(t, err, "Сервер вернул 500")

	searchMetric := metrics.Metric{
		ID:    "some_metric",
		MType: "gauge",
	}
	memStorage.GetMetric(context.TODO(), &searchMetric)
	assert.Equal(t, float64(12.0), *searchMetric.Value, "Сохраненное значение метрики типа gauge не совпадает с ожидаемым")
}

func TestPostTextCounterHandler(t *testing.T) {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	_, err := client.R().Post(srv.URL + "/update/counter/some_metric/10")
	assert.Nil(t, err, "Сервер вернул 500")

	_, err = client.R().Post(srv.URL + "/update/counter/some_metric/12")
	assert.Nil(t, err, "Сервер вернул 500")

	searchMetric := metrics.Metric{
		ID:    "some_metric",
		MType: "counter",
	}
	memStorage.GetMetric(context.TODO(), &searchMetric)
	assert.Equal(t, int64(22), *searchMetric.Delta, "Сохраненное значение метрики типа counter не совпадает с ожидаемым")
}

func TestGetTextGaugeHandler(t *testing.T) {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())
	mValue := float64(10.5)
	metric := metrics.Metric{
		ID:    "some_metric",
		MType: "gauge",
		Value: &mValue,
	}
	memStorage.SaveMetrics(context.TODO(), []metrics.Metric{metric})

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, err := client.R().Get(srv.URL + "/value/gauge/some_metric")
	assert.Nil(t, err, "Сервер вернул 500")

	body := resp.Body()
	assert.Equal(t, "10.5", string(body), "Неверное значение метрики gauge")

	resp, err = client.R().Get(srv.URL + "/value/gauge/undefined")
	assert.Nil(t, err, "Сервер вернул 500")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "Нашлась несуществующая метрика")
}

func TestGetTextCounterHandler(t *testing.T) {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())
	mValue := int64(10)
	metric := metrics.Metric{
		ID:    "some_metric",
		MType: "counter",
		Delta: &mValue,
	}
	memStorage.SaveMetrics(context.TODO(), []metrics.Metric{metric})

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

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
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	testFirstCounterValue := int64(1)
	testSecondCounterValue := int64(2)
	testFirstGaugeValue := float64(1.0)
	testSecondGaugeValue := float64(2.0)
	metricList := []metrics.Metric{
		{
			ID:    "counter_1",
			MType: "counter",
			Delta: &testFirstCounterValue,
		},
		{
			ID:    "counter_2",
			MType: "counter",
			Delta: &testSecondCounterValue,
		},
		{
			ID:    "gauge_1",
			MType: "gauge",
			Value: &testFirstGaugeValue,
		},
		{
			ID:    "gauge_2",
			MType: "gauge",
			Value: &testSecondGaugeValue,
		},
	}
	memStorage.SaveMetrics(context.TODO(), metricList)

	expectedBody := "counter_1 : 1\ncounter_2 : 2\ngauge_1 : 1\ngauge_2 : 2\n"

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

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
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

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
		requestBody  *metrics.Metric
		expectedCode int
	}{
		{
			testName:    "ok post gauge",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metric{
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
			requestBody: &metrics.Metric{
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
			requestBody: &metrics.Metric{
				ID:    "test",
				MType: "gauge",
			},
			expectedCode: http.StatusOK,
		},
		{
			testName:    "ok get counter",
			method:      http.MethodPost,
			requestPath: "/value/",
			requestBody: &metrics.Metric{
				ID:    "test",
				MType: "counter",
			},
			expectedCode: http.StatusOK,
		},
		{
			testName:    "not found get gauge",
			method:      http.MethodPost,
			requestPath: "/value/",
			requestBody: &metrics.Metric{
				ID:    "invalid",
				MType: "gauge",
			},
			expectedCode: http.StatusNotFound,
		},
		{
			testName:    "not found get counter",
			method:      http.MethodPost,
			requestPath: "/value/",
			requestBody: &metrics.Metric{
				ID:    "invalid",
				MType: "counter",
			},
			expectedCode: http.StatusNotFound,
		},
		{
			testName:    "invalid type for post",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metric{
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
			requestBody: &metrics.Metric{
				ID:    "test",
				MType: "invalid",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			testName:    "bad value for post gauge",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metric{
				ID:    "test",
				MType: "gauge",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			testName:    "bad value for post counter",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metric{
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

func TestInvalidStorageJSONHandlers(t *testing.T) {
	fileWorker := fileworker.New("", &InvalidStorage{})
	handlers := New(&config.Config{StoreInterval: 1}, &InvalidStorage{}, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()

	defaultValue := float64(5.5)
	testCases := []struct {
		testName     string
		method       string
		requestPath  string
		requestBody  *metrics.Metric
		expectedCode int
	}{
		{
			testName:    "test error for post",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metric{
				ID:    "test",
				MType: "gauge",
				Value: &defaultValue,
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			testName:    "test error for get",
			method:      http.MethodPost,
			requestPath: "/value/",
			requestBody: &metrics.Metric{
				ID:    "test",
				MType: "gauge",
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			req, _ := json.Marshal(tc.requestBody)
			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(req).
				Execute(tc.method, srv.URL+tc.requestPath)
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestJsonGaugeStorageHandler(t *testing.T) {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()

	val := float64(5.0)
	metric := &metrics.Metric{
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

func TestPostSeveralHandler(t *testing.T) {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	gaugeValue := float64(10.0)
	counterValue := int64(10)
	metricList := []metrics.Metric{
		{
			ID:    "test",
			MType: "gauge",
			Value: &gaugeValue,
		},
		{
			ID:    "test",
			MType: "counter",
			Delta: &counterValue,
		},
	}
	req, _ := json.Marshal(metricList)

	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(srv.URL + "/updates/")
	assert.Nil(t, err, "Сервер вернул 500")

	searchCounterMetric := metrics.Metric{
		ID:    "test",
		MType: "counter",
	}
	searchGaugeMetric := metrics.Metric{
		ID:    "test",
		MType: "gauge",
	}
	memStorage.GetMetric(context.TODO(), &searchCounterMetric)
	assert.Equal(t, int64(10), *searchCounterMetric.Delta, "Сохраненное значение метрики типа counter не совпадает с ожидаемым")
	memStorage.GetMetric(context.TODO(), &searchGaugeMetric)
	assert.Equal(t, float64(10.0), *searchGaugeMetric.Value, "Сохраненное значение метрики типа gauge не совпадает с ожидаемым")
}

func TestPostSeveralBadRequestHandler(t *testing.T) {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody("invalid").
		Post(srv.URL + "/updates/")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Статус не совпадает")

	resp, _ = client.R().
		SetHeader("Content-Type", "application/json").
		Post(srv.URL + "/updates/")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Статус не совпадает")

	resp, _ = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]byte("[{\"type\": \"INVALID\"}]")).
		Post(srv.URL + "/updates/")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Статус не совпадает")
}

func TestPostJsonBadRequestHandler(t *testing.T) {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody("invalid").
		Post(srv.URL + "/update/")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Статус не совпадает")

	resp, _ = client.R().
		SetHeader("Content-Type", "application/json").
		Post(srv.URL + "/update/")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Статус не совпадает")

	resp, _ = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]byte("{\"type\": \"INVALID\"}")).
		Post(srv.URL + "/update/")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Статус не совпадает")
}

func TestGetJsonBadRequestHandler(t *testing.T) {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 0}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody("invalid").
		Post(srv.URL + "/value/")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Статус не совпадает")

	resp, _ = client.R().
		SetHeader("Content-Type", "application/json").
		Post(srv.URL + "/value/")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Статус не совпадает")

	resp, _ = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]byte("{\"type\": \"INVALID\"}")).
		Post(srv.URL + "/value/")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Статус не совпадает")

	resp, _ = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]byte("{\"type\": \"gauge\", \"id\": \"invalid\"}")).
		Post(srv.URL + "/value/")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "Статус не совпадает")
}

func TestInvalidStoragePlainHandlers(t *testing.T) {
	fileWorker := fileworker.New("", &InvalidStorage{})
	handlers := New(&config.Config{StoreInterval: 1}, &InvalidStorage{}, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()

	defaultValue := float64(5.5)
	defaultDelta := int64(5)
	testCases := []struct {
		testName     string
		method       string
		requestPath  string
		requestBody  *metrics.Metric
		expectedCode int
	}{
		{
			testName:    "test error for post gauge",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metric{
				ID:    "test",
				MType: "gauge",
				Value: &defaultValue,
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			testName:    "test error for post counter",
			method:      http.MethodPost,
			requestPath: "/update/",
			requestBody: &metrics.Metric{
				ID:    "test",
				MType: "counter",
				Delta: &defaultDelta,
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			testName:    "test error for get gauge",
			method:      http.MethodGet,
			requestPath: "/value/",
			requestBody: &metrics.Metric{
				ID:    "test",
				MType: "gauge",
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			testName:    "test error for get counter",
			method:      http.MethodGet,
			requestPath: "/value/",
			requestBody: &metrics.Metric{
				ID:    "test",
				MType: "counter",
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			reqURL := srv.URL + tc.requestPath + tc.requestBody.MType + "/" + tc.requestBody.ID
			if tc.requestBody.Delta != nil {
				reqURL += "/" + strconv.FormatInt(*tc.requestBody.Delta, 10)
			}
			if tc.requestBody.Value != nil {
				reqURL += "/" + strconv.FormatFloat(*tc.requestBody.Value, 'f', -1, 64)
			}
			resp, _ := client.R().Execute(tc.method, reqURL)
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestGetAllInvalidStorageHandler(t *testing.T) {
	fileWorker := fileworker.New("", &InvalidStorage{})
	handlers := New(&config.Config{StoreInterval: 0}, &InvalidStorage{}, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, _ := client.R().Get(srv.URL + "/")
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode(), "Статус не совпадает")
}
