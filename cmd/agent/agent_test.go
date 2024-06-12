package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/internal/storage"
)

type MockServerStorage struct {
	PathTimesCalled map[string]int64
}

func (mStorage *MockServerStorage) pathCounter(res http.ResponseWriter, req *http.Request) {
	if mStorage.PathTimesCalled == nil {
		mStorage.PathTimesCalled = map[string]int64{}
	}
	mStorage.PathTimesCalled[req.URL.Path] += 1
}

func (mStorage *MockServerStorage) mockRouter() chi.Router {
	router := chi.NewRouter()
    router.Get("/", mStorage.pathCounter)
    return router
}

func TestCollectMetric(t *testing.T) {
	storage := storage.MetricStorage{}
	CollectMetric(&storage)

	storedGaugeValues := storage.GetGaugeValues()

	assert.Equal(t, 28, len(storedGaugeValues), "Несовпадает количество отслеживаемых метрик")
	assert.Contains(t, storedGaugeValues, "NextGC")
	assert.Contains(t, storedGaugeValues, "HeapIdle")
	assert.Contains(t, storedGaugeValues, "RandomValue")
	assert.Equal(t, int64(1), storage.GetCounterValue("PollCount"))

	CollectMetric(&storage)

	assert.Equal(t, int64(2), storage.GetCounterValue("PollCount"))
}

func TestSendMetric(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	agentStorage := storage.MetricStorage{}

	agentStorage.UpdateGaugeValue("test", 10.0)
	agentStorage.UpdateGaugeValue("test_2", 10.0)
	agentStorage.UpdateGaugeValue("test_3", 10.0)
	agentStorage.UpdateCounterValue("test_4", 10)
	agentStorage.UpdateCounterValue("test_5", 7)

	SendMetric(&agentStorage, srv.URL)
	assert.Equal(t, 5, len(serverStorage.PathTimesCalled), "Не прошел запрос на сервер")
	assert.Equal(t, int64(1), serverStorage.PathTimesCalled["/update/gauge/test/10"], "Неверный запрос сервера")
}

func TestRun(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	agentStorage := storage.MetricStorage{}
	agentStorage.UpdateCounterValue("PollCount", 3)

	Run(&agentStorage, srv.URL)
	// PollCount == 4
	assert.Nil(t, serverStorage.PathTimesCalled, "Вызвался сервер, когда PollCount не кратен 5")
	Run(&agentStorage, srv.URL)
	// PollCount == 5 => server called
	assert.Less(t, 0, len(serverStorage.PathTimesCalled), "Не прошел запрос на сервер")
}
