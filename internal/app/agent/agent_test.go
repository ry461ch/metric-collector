package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/app/agent/config"
	"github.com/ry461ch/metric-collector/internal/models/netaddr"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/metricservice"
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
	metric := metrics.Metrics{}
	json.Unmarshal(buf.Bytes(), &metric)

	if m.metricsCounter == nil {
		m.metricsCounter = map[string]int64{}
	}
	if m.metricsGauge == nil {
		m.metricsGauge = map[string]float64{}
	}

	switch metric.MType {
	case "gauge":
		m.metricsGauge[metric.ID] = *metric.Value
	case "counter":
		m.metricsCounter[metric.ID] = *metric.Delta
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
	metricStorage := memstorage.MemStorage{}
	metricService := metricservice.New(&metricStorage)
	agent := New(&TimeState{}, &config.Config{}, metricService)
	agent.collectMetric()

	storedGaugeValues := metricStorage.GetGaugeValues()

	assert.Equal(t, 28, len(storedGaugeValues), "Несовпадает количество отслеживаемых метрик")
	assert.Contains(t, storedGaugeValues, "NextGC")
	assert.Contains(t, storedGaugeValues, "HeapIdle")
	assert.Contains(t, storedGaugeValues, "RandomValue")
	val, _ := metricStorage.GetCounterValue("PollCount")
	assert.Equal(t, int64(1), val)

	agent.collectMetric()

	val, _ = metricStorage.GetCounterValue("PollCount")
	assert.Equal(t, int64(2), val)
}

func TestSendMetric(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	agentStorage := memstorage.MemStorage{}

	agentStorage.UpdateGaugeValue("test_1", 10.0)
	agentStorage.UpdateGaugeValue("test_2", 10.0)
	agentStorage.UpdateGaugeValue("test_3", 10.0)
	agentStorage.UpdateCounterValue("test_4", 10)
	agentStorage.UpdateCounterValue("test_5", 7)

	metricService := metricservice.New(&agentStorage)
	agent := New(&TimeState{}, &config.Config{Addr: splitURL(srv.URL)}, metricService)

	agent.sendMetrics()
	assert.Equal(t, int64(5), serverStorage.timesCalled, "Не прошел запрос на сервер")
	assert.Equal(t, float64(10.0), serverStorage.metricsGauge["test_2"], "Неправильно записалась метрика в хранилище")
	assert.Equal(t, int64(10), serverStorage.metricsCounter["test_4"], "Неправильно записалась метрика в хранилище")
}

func TestRun(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	agentStorage := memstorage.MemStorage{}
	config := config.Config{ReportIntervalSec: 6, PollIntervalSec: 3, Addr: splitURL(srv.URL)}
	timeState := TimeState{LastCollectMetricTime: time.Now(), LastSendMetricTime: time.Now()}

	metricService := metricservice.New(&agentStorage)
	agent := New(&timeState, &config, metricService)

	agent.runIteration()
	pollCount, _ := agentStorage.GetCounterValue("PollCount")
	assert.Equal(t, int64(0), pollCount, "Вызвался collect metric, хотя еще не должен был")

	timeState.LastCollectMetricTime = time.Now().Add(-time.Second * 4)
	agent.runIteration()
	pollCount, _ = agentStorage.GetCounterValue("PollCount")
	assert.Equal(t, int64(1), pollCount, "Кол-во вызовов collectMetric не совпадает с ожидаемым")

	timeState.LastSendMetricTime = time.Now().Add(-time.Second * 7)
	agent.runIteration()
	assert.Less(t, int64(0), serverStorage.timesCalled, "Не прошел запрос на сервер")
	pollCount, _ = agentStorage.GetCounterValue("PollCount")
	assert.Equal(t, int64(1), pollCount, "Кол-во вызовов collectMetric не совпадает с ожидаемым")
}
