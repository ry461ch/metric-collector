package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"
	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

func mockRouter(handlers *Handlers) chi.Router {
	router := chi.NewRouter()
	router.Post("/update/counter/{name}/{value}", handlers.PostCounterHandler)
	router.Post("/update/gauge/{name}/{value}", handlers.PostGaugeHandler)
	router.Get("/value/counter/{name}", handlers.GetCounterHandler)
	router.Get("/value/gauge/{name}", handlers.GetGaugeHandler)
	router.Get("/", handlers.GetAllMetricsHandler)
	return router
}

func TestPostGaugeServe(t *testing.T) {
	memStorage := memstorage.MemStorage{}

	handlers := Handlers{mStorage: &memStorage}
	router := mockRouter(&handlers)
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

func TestPostCounterServe(t *testing.T) {
	memStorage := memstorage.MemStorage{}

	handlers := Handlers{mStorage: &memStorage}
	router := mockRouter(&handlers)
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

func TestGetGaugeServe(t *testing.T) {
	memStorage := memstorage.MemStorage{}
	memStorage.UpdateGaugeValue("some_metric", 10.5)

	handlers := Handlers{mStorage: &memStorage}
	router := mockRouter(&handlers)
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

func TestGetCounterServe(t *testing.T) {
	memStorage := memstorage.MemStorage{}
	memStorage.UpdateCounterValue("some_metric", 10)

	handlers := Handlers{mStorage: &memStorage}
	router := mockRouter(&handlers)
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

func TestGetAllMetrics(t *testing.T) {
	memStorage := memstorage.MemStorage{}
	memStorage.UpdateCounterValue("counter_1", 1)
	memStorage.UpdateCounterValue("counter_2", 2)
	memStorage.UpdateGaugeValue("gauge_1", 1)
	memStorage.UpdateGaugeValue("gauge_2", 2)

	expectedBody := "counter_1 : 1\ncounter_2 : 2\ngauge_1 : 1\ngauge_2 : 2\n"

	handlers := Handlers{mStorage: &memStorage}
	router := mockRouter(&handlers)
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
