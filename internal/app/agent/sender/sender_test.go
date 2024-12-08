package sender

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	config "github.com/ry461ch/metric-collector/internal/config/agent"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/internal/models/netaddr"
	"github.com/ry461ch/metric-collector/pkg/encrypt"
	rsacomponent "github.com/ry461ch/metric-collector/pkg/rsa"
	"github.com/ry461ch/metric-collector/pkg/rsa/middleware"
)

type MockServerStorage struct {
	timesCalledMutex sync.Mutex
	timesCalled      int64
	gaugeMutex       sync.Mutex
	metricsGauge     map[string]float64
	counterMutex     sync.Mutex
	metricsCounter   map[string]int64
}

func (m *MockServerStorage) handler(res http.ResponseWriter, req *http.Request) {
	m.timesCalledMutex.Lock()
	m.timesCalled += 1
	m.timesCalledMutex.Unlock()

	var buf bytes.Buffer
	buf.ReadFrom(req.Body)
	metricList := []metrics.Metric{}
	json.Unmarshal(buf.Bytes(), &metricList)

	m.counterMutex.Lock()
	if m.metricsCounter == nil {
		m.metricsCounter = map[string]int64{}
	}
	m.counterMutex.Unlock()

	m.gaugeMutex.Lock()
	if m.metricsGauge == nil {
		m.metricsGauge = map[string]float64{}
	}
	m.gaugeMutex.Unlock()

	for _, metric := range metricList {
		switch metric.MType {
		case "gauge":
			m.gaugeMutex.Lock()
			m.metricsGauge[metric.ID] = *metric.Value
			m.gaugeMutex.Unlock()
		case "counter":
			m.counterMutex.Lock()
			m.metricsCounter[metric.ID] = *metric.Delta
			m.counterMutex.Unlock()
		}
	}
	res.WriteHeader(http.StatusOK)
}

func (m *MockServerStorage) mockRouter(decrypter *rsacomponent.RsaDecrypter) chi.Router {
	router := chi.NewRouter()
	if decrypter != nil {
		router.Use(rsamiddleware.DecryptRequest(decrypter))
	}
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
	privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	publicKey := privateKey.PublicKey

	privateKeyPath := "/tmp/private.test"
	publicKeyPath := "/tmp/public.test"

	privateKeyFile, _ := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer privateKeyFile.Close()

	privateKeyBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
	privateKeyFile.Write(privateKeyBytes)

	publicKeyFile, _ := os.OpenFile(publicKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer publicKeyFile.Close()

	publicKeyBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&publicKey),
		},
	)
	publicKeyFile.Write(publicKeyBytes)

	encrypter := rsacomponent.NewEncrypter(publicKeyPath)
	encrypter.Initialize(context.TODO())
	decrypter := rsacomponent.NewDecrypter(privateKeyPath)
	decrypter.Initialize(context.TODO())

	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter(decrypter)
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

	sender := New(
		encrypt.New("test"),
		encrypter,
		&config.Config{Addr: *splitURL(srv.URL), RateLimit: 2, ReportIntervalSec: 1},
		"127.0.0.1",
	)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1)
	defer cancel()
	go sender.Run(ctx, metricChannel)
	time.Sleep(time.Second)

	assert.Equal(t, int64(5), serverStorage.timesCalled, "Не прошел запрос на сервер")
	assert.Equal(t, float64(10.0), serverStorage.metricsGauge["test_3"], "Неправильно записалась метрика в хранилище")
	assert.Equal(t, int64(10), serverStorage.metricsCounter["test_1"], "Неправильно записалась метрика в хранилище")
}

func BenchmarkSendMetric(b *testing.B) {
	testCounterValue := int64(10)
	testGaugeValue := float64(10.0)

	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter(nil)
	srv := httptest.NewServer(router)
	defer srv.Close()

	metricChannel := make(chan metrics.Metric, 100)
	defer close(metricChannel)

	sender := New(encrypt.New("test"), nil, &config.Config{Addr: *splitURL(srv.URL), RateLimit: 2}, "127.0.0.1")

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for j := 0; j < 50; j++ {
			metricChannel <- metrics.Metric{
				ID:    "test_1",
				MType: "counter",
				Delta: &testCounterValue,
			}
			metricChannel <- metrics.Metric{
				ID:    "test_2",
				MType: "gauge",
				Value: &testGaugeValue,
			}
		}

		b.StartTimer()
		sender.sendMetrics(context.TODO(), metricChannel)
	}
}
