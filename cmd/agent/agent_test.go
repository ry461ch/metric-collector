package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/metric-collector/internal/storage"
)


type MockClient struct {
	PathTimesCalled map[string]int64
}

func (mock_client *MockClient) Post(path string) (int64, error) {
	if mock_client.PathTimesCalled == nil {
		mock_client.PathTimesCalled = map[string]int64{}
	}
	mock_client.PathTimesCalled[path] += 1
	return 200, nil
}

func TestCollectMetric(t *testing.T) {
	storage := storage.MetricStorage{}
	CollectMetric(&storage)

	stored_gauge_values := storage.GetGaugeValues()
	
	assert.Equal(t, 28, len(stored_gauge_values), "Несовпадает количество отслеживаемых метрик")
	assert.Contains(t, stored_gauge_values, "NextGC")
	assert.Contains(t, stored_gauge_values, "HeapIdle")
	assert.Contains(t, stored_gauge_values, "RandomValue")
	assert.Equal(t, int64(1), storage.GetCounterValue("PollCount"))
	
	CollectMetric(&storage)
	
	assert.Equal(t, int64(2), storage.GetCounterValue("PollCount"))
}

func TestSendMetric(t *testing.T) {
	client := MockClient{}
	storage := storage.MetricStorage{}
	
	storage.UpdateGaugeValue("test", 10.0)
	storage.UpdateGaugeValue("test_2", 10.0)
	storage.UpdateGaugeValue("test_3", 10.0)
	storage.UpdateCounterValue("test_4", 10)
	storage.UpdateCounterValue("test_5", 7)
	
	SendMetric(&storage, &client)
	assert.Equal(t, 5, len(client.PathTimesCalled), "Не прошел запрос на сервер")
	assert.Equal(t, int64(1), client.PathTimesCalled["/update/gauge/test/10"], "Неверный запрос сервера")
}

func TestRun(t *testing.T) {
	client := MockClient{}
	storage := storage.MetricStorage{}
	storage.UpdateCounterValue("PollCount", 3)
	
	Run(&storage, &client)
	// PollCount == 4
	assert.Nil(t, client.PathTimesCalled, "Вызвался сервер, когда PollCount не кратен 5")
	Run(&storage, &client)
	// PollCount == 5 => server called
	assert.Less(t, 0, len(client.PathTimesCalled), "Не прошел запрос на сервер")
}
