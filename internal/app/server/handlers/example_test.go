package handlers

import (
	"context"
	"fmt"
	"net/http/httptest"

	"gopkg.in/resty.v1"

	config "github.com/ry461ch/metric-collector/internal/config/server"
	"github.com/ry461ch/metric-collector/internal/fileworker"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	memstorage "github.com/ry461ch/metric-collector/internal/storage/memory"
)

func ExampleHandlers_PostPlainGaugeHandler() {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 1}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	res, _ := client.R().Post(srv.URL + "/update/gauge/some_metric/10.0")
	fmt.Println(res.StatusCode())

	// Output:
	// 200
}

func ExampleHandlers_PostPlainCounterHandler() {
	memStorage := memstorage.New()
	memStorage.Initialize(context.TODO())

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 1}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	res, _ := client.R().Post(srv.URL + "/update/counter/some_metric/10")
	fmt.Println(res.StatusCode())

	// Output:
	// 200
}

func ExampleHandlers_GetPlainGaugeHandler() {
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
	handlers := New(&config.Config{StoreInterval: 1}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, _ := client.R().Get(srv.URL + "/value/gauge/some_metric")
	fmt.Println(resp.StatusCode())

	body := resp.Body()
	fmt.Println(string(body))

	// Output:
	// 200
	// 10.5
}

func ExampleHandlers_GetPlainCounterHandler() {
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
	handlers := New(&config.Config{StoreInterval: 1}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, _ := client.R().Get(srv.URL + "/value/counter/some_metric")
	fmt.Println(resp.StatusCode())

	body := resp.Body()
	fmt.Println(string(body))

	// Output:
	// 200
	// 10
}

func ExampleHandlers_GetPlainAllMetricsHandler() {
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

	fileWorker := fileworker.New("", memStorage)
	handlers := New(&config.Config{StoreInterval: 1}, memStorage, fileWorker)

	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, _ := client.R().Get(srv.URL + "/")
	fmt.Println(resp.StatusCode())

	body := resp.Body()
	fmt.Println(len(string(body)))

	// Output:
	// 200
	// 52
}
