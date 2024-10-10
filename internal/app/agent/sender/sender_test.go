package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	config "github.com/ry461ch/metric-collector/internal/config/agent"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/models/netaddr"
	"github.com/ry461ch/metric-collector/pkg/encrypt"
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
	res.WriteHeader(http.StatusOK)
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

func TestSendMetric(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

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

	metricChannel := make(chan metrics.Metric, 5)
	defer close(metricChannel)
	for _, metric := range metricList {
		metricChannel <- metric
	}

	sender := Sender{
		cfg:       &config.Config{Addr: *splitURL(srv.URL), RateLimit: 2},
		encrypter: encrypt.New("test"),
	}

	sender.sendMetrics(context.TODO(), metricChannel)
	assert.Equal(t, int64(5), serverStorage.timesCalled, "Не прошел запрос на сервер")
	assert.Equal(t, float64(10.0), serverStorage.metricsGauge["test_3"], "Неправильно записалась метрика в хранилище")
	assert.Equal(t, int64(10), serverStorage.metricsCounter["test_1"], "Неправильно записалась метрика в хранилище")
}
