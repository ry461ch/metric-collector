package agent

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/agent/config"
	"github.com/ry461ch/metric-collector/internal/models/netaddr"
	"github.com/ry461ch/metric-collector/internal/storage/memory"
)

type MockServerStorage struct {
	PathTimesCalled map[string]int64
}

func (m *MockServerStorage) pathCounter(res http.ResponseWriter, req *http.Request) {
	if m.PathTimesCalled == nil {
		m.PathTimesCalled = map[string]int64{}
	}

	m.PathTimesCalled[req.URL.Path] += 1
}

func (m *MockServerStorage) mockRouter() chi.Router {
	router := chi.NewRouter()
	router.Post("/*", m.pathCounter)
	return router
}

func splitURL(URL string) netaddr.NetAddress {
	updatedURL, _ := strings.CutPrefix(URL, "http://")
	parts := strings.Split(updatedURL, ":")
	port, _ := strconv.ParseInt(parts[1], 10, 0)
	return netaddr.NetAddress{Host: parts[0], Port: port}
}

func TestCollectMetric(t *testing.T) {
	storage := memstorage.MemStorage{}
	agent := Agent{TimeState: &config.TimeState{}, Options: config.Options{}, MStorage: &storage}
	agent.CollectMetric()

	storedGaugeValues := storage.GetGaugeValues()

	assert.Equal(t, 28, len(storedGaugeValues), "Несовпадает количество отслеживаемых метрик")
	assert.Contains(t, storedGaugeValues, "NextGC")
	assert.Contains(t, storedGaugeValues, "HeapIdle")
	assert.Contains(t, storedGaugeValues, "RandomValue")
	val, _ := storage.GetCounterValue("PollCount")
	assert.Equal(t, int64(1), val)

	agent.CollectMetric()

	val, _ = storage.GetCounterValue("PollCount")
	assert.Equal(t, int64(2), val)
}

func TestSendMetric(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	agentStorage := memstorage.MemStorage{}

	agentStorage.UpdateGaugeValue("test", 10.0)
	agentStorage.UpdateGaugeValue("test_2", 10.0)
	agentStorage.UpdateGaugeValue("test_3", 10.0)
	agentStorage.UpdateCounterValue("test_4", 10)
	agentStorage.UpdateCounterValue("test_5", 7)

	agent := Agent{TimeState: &config.TimeState{}, Options: config.Options{Addr: splitURL(srv.URL)}, MStorage: &agentStorage}

	agent.SendMetric()
	assert.Equal(t, 5, len(serverStorage.PathTimesCalled), "Не прошел запрос на сервер")
	assert.Equal(t, int64(1), serverStorage.PathTimesCalled["/update/gauge/test/10"], "Неверный запрос серверу")
}

func TestRun(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	agentStorage := memstorage.MemStorage{}
	options := config.Options{ReportIntervalSec: 6, PollIntervalSec: 3, Addr: splitURL(srv.URL)}
	timeState := config.TimeState{LastCollectMetricTime: time.Now(), LastSendMetricTime: time.Now()}
	agent := Agent{TimeState: &timeState, Options: options, MStorage: &agentStorage}

	agent.Run()
	assert.Nil(t, serverStorage.PathTimesCalled, "Вызвался сервер, хотя еще не должен был")
	pollCount, _ := agentStorage.GetCounterValue("PollCount")
	assert.Equal(t, int64(0), pollCount, "Вызвался collect metric, хотя еще не должен был")

	timeState.LastCollectMetricTime = time.Now().Add(-time.Second * 4)
	agent.Run()
	assert.Nil(t, serverStorage.PathTimesCalled, "Вызвался сервер, хотя еще не должен был")
	pollCount, _ = agentStorage.GetCounterValue("PollCount")
	assert.Equal(t, int64(1), pollCount, "Кол-во вызовов collectMetric не совпадает с ожидаемым")

	timeState.LastSendMetricTime = time.Now().Add(-time.Second * 7)
	agent.Run()
	assert.Less(t, 0, len(serverStorage.PathTimesCalled), "Не прошел запрос на сервер")
	pollCount, _ = agentStorage.GetCounterValue("PollCount")
	assert.Equal(t, int64(1), pollCount, "Кол-во вызовов collectMetric не совпадает с ожидаемым")
}
