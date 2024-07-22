package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	config "github.com/ry461ch/metric-collector/internal/config/agent"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/models/netaddr"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

type MockServerStorage struct {
	timesCalled    int64
	metricsGauge   map[string]float64
	metricsCounter map[string]int64
}

func (m *MockServerStorage) handler(res http.ResponseWriter, req *http.Request) {
	m.timesCalled += 1

	var buf bytes.Buffer
	buf.ReadFrom(req.Body)
	metricList := []metrics.Metric{}
	json.Unmarshal(buf.Bytes(), &metricList)

	if m.metricsCounter == nil {
		m.metricsCounter = map[string]int64{}
	}
	if m.metricsGauge == nil {
		m.metricsGauge = map[string]float64{}
	}

	for _, metric := range metricList {
		switch metric.MType {
		case "gauge":
			m.metricsGauge[metric.ID] = *metric.Value
		case "counter":
			m.metricsCounter[metric.ID] = *metric.Delta
		}
	}
}

func (m *MockServerStorage) mockRouter() chi.Router {
	router := chi.NewRouter()
	router.Post("/*", m.handler)
	return router
}

func splitURL(URL string) *netaddr.NetAddress {
	updatedURL, _ := strings.CutPrefix(URL, "http://")
	parts := strings.Split(updatedURL, ":")
	port, _ := strconv.ParseInt(parts[1], 10, 0)
	return &netaddr.NetAddress{Host: parts[0], Port: port}
}

func TestCollectMetric(t *testing.T) {
	metricStorage := memstorage.NewMemStorage()
	metricStorage.Initialize(context.TODO())
	agent := Agent{
		timeState:  &TimeState{},
		config:     &config.Config{},
		memStorage: metricStorage,
	}
	agent.collectMetric(context.TODO())

	storedMetrics, _ := metricStorage.ExtractMetrics(context.TODO())

	assert.Equal(t, 29, len(storedMetrics), "Несовпадает количество отслеживаемых метрик")

	agent.collectMetric(context.TODO())

	searchMetric := metrics.Metric{
		ID:    "PollCount",
		MType: "counter",
	}
	metricStorage.GetMetric(context.TODO(), &searchMetric)
	assert.Equal(t, int64(2), *searchMetric.Delta)
}

func TestSendMetric(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	agentStorage := memstorage.NewMemStorage()
	agentStorage.Initialize(context.TODO())

	testFirstCounterValue := int64(10)
	testSecondCounterValue := int64(7)
	testFirstGaugeValue := float64(10.0)
	testSecondGaugeValue := float64(7.0)
	testThirdGaugeValue := float64(7.0)
	metricList := []metrics.Metric{
		{
			ID:    "test_1",
			MType: "counter",
			Delta: &testFirstCounterValue,
		},
		{
			ID:    "test_2",
			MType: "counter",
			Delta: &testSecondCounterValue,
		},
		{
			ID:    "test_3",
			MType: "gauge",
			Value: &testFirstGaugeValue,
		},
		{
			ID:    "test_4",
			MType: "gauge",
			Value: &testSecondGaugeValue,
		},
		{
			ID:    "test_5",
			MType: "gauge",
			Value: &testThirdGaugeValue,
		},
	}
	agentStorage.SaveMetrics(context.TODO(), metricList)

	agent := Agent{
		timeState:  &TimeState{},
		config:     &config.Config{Addr: *splitURL(srv.URL)},
		memStorage: agentStorage,
	}

	agent.sendMetrics(context.TODO())
	assert.Equal(t, int64(1), serverStorage.timesCalled, "Не прошел запрос на сервер")
	assert.Equal(t, float64(10.0), serverStorage.metricsGauge["test_3"], "Неправильно записалась метрика в хранилище")
	assert.Equal(t, int64(10), serverStorage.metricsCounter["test_1"], "Неправильно записалась метрика в хранилище")
}

func TestRun(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	agentStorage := memstorage.NewMemStorage()
	agentStorage.Initialize(context.TODO())
	config := config.Config{ReportIntervalSec: 6, PollIntervalSec: 3, Addr: *splitURL(srv.URL)}
	timeState := TimeState{LastCollectMetricTime: time.Now(), LastSendMetricTime: time.Now()}

	agent := Agent{
		timeState:  &timeState,
		config:     &config,
		memStorage: agentStorage,
	}

	agent.collectAndSendMetrics(context.TODO())

	searchMetric := metrics.Metric{
		ID:    "PollCount",
		MType: "counter",
	}
	agentStorage.GetMetric(context.TODO(), &searchMetric)
	assert.Nil(t, searchMetric.Delta, "Вызвался collect metric, хотя еще не должен был")

	timeState.LastCollectMetricTime = time.Now().Add(-time.Second * 4)
	agent.collectAndSendMetrics(context.TODO())
	agentStorage.GetMetric(context.TODO(), &searchMetric)
	assert.Equal(t, int64(1), *searchMetric.Delta, "Кол-во вызовов collectMetric не совпадает с ожидаемым")

	timeState.LastSendMetricTime = time.Now().Add(-time.Second * 7)
	agent.collectAndSendMetrics(context.TODO())
	assert.Less(t, int64(0), serverStorage.timesCalled, "Не прошел запрос на сервер")
	agentStorage.GetMetric(context.TODO(), &searchMetric)
	assert.Equal(t, int64(1), *searchMetric.Delta, "Кол-во вызовов collectMetric не совпадает с ожидаемым")
}
