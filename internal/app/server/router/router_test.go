package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"

	"github.com/ry461ch/metric-collector/pkg/encrypt"
	"github.com/ry461ch/metric-collector/pkg/logging"
)

type MockHandlers struct {
	pathTimesCalled map[string]int64
}

func NewMockHandlers() MockHandlers {
	return MockHandlers{pathTimesCalled: map[string]int64{}}
}

func (m *MockHandlers) PostPlainGaugeHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["postGauge"] += 1
	res.WriteHeader(http.StatusOK)
}

func (m *MockHandlers) PostPlainCounterHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["postCounter"] += 1
	res.WriteHeader(http.StatusOK)
}

func (m *MockHandlers) GetPlainGaugeHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["getGauge"] += 1
	res.WriteHeader(http.StatusOK)
}

func (m *MockHandlers) GetPlainCounterHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["getCounter"] += 1
	res.WriteHeader(http.StatusOK)
}

func (m *MockHandlers) GetPlainAllMetricsHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["getAll"] += 1
	res.WriteHeader(http.StatusOK)
}

func (m *MockHandlers) PostJSONHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["postJson"] += 1
	res.WriteHeader(http.StatusOK)
}

func (m *MockHandlers) GetJSONHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["getJson"] += 1
	res.WriteHeader(http.StatusOK)
}

func (m *MockHandlers) PostMetricsHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["postAllJson"] += 1
	res.WriteHeader(http.StatusOK)
}

func (m *MockHandlers) Ping(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["ping"] += 1
	res.WriteHeader(http.StatusOK)
}

func TestRouter(t *testing.T) {
	defaultPostGaugeRequest := "/update/gauge/some_metric/10.0"
	defaultPostCounterRequest := "/update/counter/some_metric/10"
	defaultGetGaugeRequest := "/value/gauge/some_metric"
	defaultGetCounterRequest := "/value/counter/some_metric"

	jsonContentType := "application/json"
	plainContentType := "text/plain"

	testCases := []struct {
		testName                string
		method                  string
		requestPath             string
		requestContentType      string
		expectedCode            int
		expectedPathTimesCalled map[string]int64
	}{
		{
			testName:                "invalid method",
			method:                  http.MethodDelete,
			requestPath:             defaultPostCounterRequest,
			requestContentType:      plainContentType,
			expectedCode:            http.StatusMethodNotAllowed,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "ok for post gauge",
			method:                  http.MethodPost,
			requestPath:             defaultPostGaugeRequest,
			requestContentType:      plainContentType,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"postGauge": 1},
		},
		{
			testName:                "ok for post counter",
			method:                  http.MethodPost,
			requestPath:             defaultPostCounterRequest,
			requestContentType:      plainContentType,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"postCounter": 1},
		},
		{
			testName:                "ok for get counter",
			method:                  http.MethodGet,
			requestPath:             defaultGetCounterRequest,
			requestContentType:      plainContentType,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"getCounter": 1},
		},
		{
			testName:                "ok for get gauge",
			method:                  http.MethodGet,
			requestPath:             defaultGetGaugeRequest,
			requestContentType:      plainContentType,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"getGauge": 1},
		},
		{
			testName:                "ok for get all",
			method:                  http.MethodGet,
			requestPath:             "/",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"getAll": 1},
		},
		{
			testName:                "no metric name and value for post counter",
			method:                  http.MethodPost,
			requestPath:             "/update/counter/",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "no metric name and value for post gauge",
			method:                  http.MethodPost,
			requestPath:             "/update/gauge/",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid metric type for post",
			method:                  http.MethodPost,
			requestPath:             "/update/invalid_metric_type",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "no metric name for post counter",
			method:                  http.MethodPost,
			requestPath:             "/update/counter/10",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "no metric name for post gauge",
			method:                  http.MethodPost,
			requestPath:             "/update/gauge/10",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid metric value for post counter",
			method:                  http.MethodPost,
			requestPath:             "/update/counter/test/10.0",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid metric value for post gauge",
			method:                  http.MethodPost,
			requestPath:             "/update/gauge/test/str",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "no metric name for get counter",
			method:                  http.MethodGet,
			requestPath:             "/value/counter/",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "no metric name for get gauge",
			method:                  http.MethodGet,
			requestPath:             "/value/gauge/",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "no metric type for get",
			method:                  http.MethodGet,
			requestPath:             "/value/invalid/",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "bad path for get",
			method:                  http.MethodGet,
			requestPath:             "/invalid",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid content-type for post text/plain",
			method:                  http.MethodPost,
			requestPath:             defaultPostCounterRequest,
			requestContentType:      jsonContentType,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid content-type for get text/plain",
			method:                  http.MethodGet,
			requestPath:             defaultGetCounterRequest,
			requestContentType:      jsonContentType,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid content-type for post application/json",
			method:                  http.MethodPost,
			requestPath:             "/update/",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "invalid content-type for get application/json",
			method:                  http.MethodPost,
			requestPath:             "/value/",
			requestContentType:      plainContentType,
			expectedCode:            http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName:                "ok for post application/json",
			method:                  http.MethodPost,
			requestPath:             "/update/",
			requestContentType:      jsonContentType,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"postJson": 1},
		},
		{
			testName:                "ok for get application/json",
			method:                  http.MethodPost,
			requestPath:             "/value/",
			requestContentType:      jsonContentType,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"getJson": 1},
		},
		{
			testName:                "ok for ping",
			method:                  http.MethodGet,
			requestPath:             "/ping",
			requestContentType:      jsonContentType,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"ping": 1},
		},
		{
			testName:                "ok for post metrics",
			method:                  http.MethodPost,
			requestPath:             "/updates/",
			requestContentType:      jsonContentType,
			expectedCode:            http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"postAllJson": 1},
		},
	}
	logging.Initialize("WARN")
	logging.Initialize("ERROR")
	logging.Initialize("DEBUG")
	logging.Initialize("INVALID")
	logging.Initialize("INFO")

	client := resty.New()

	handlers := NewMockHandlers()
	encrypter := encrypt.New("test")

	router := New(&handlers, encrypter, nil)
	srv := httptest.NewServer(router)
	defer srv.Close()

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			reqBodyHash := encrypter.EncryptMessage([]byte(""))

			resp, err := client.R().
				SetHeader("Content-Type", tc.requestContentType).
				SetHeader("HashSHA256", fmt.Sprintf("%x", reqBodyHash)).
				SetBody([]byte("")).
				Execute(tc.method, srv.URL+tc.requestPath)
			assert.Nil(t, err, "Сервер вернул 500")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, len(tc.expectedPathTimesCalled), len(handlers.pathTimesCalled), "Запрос прошел до сервиса, хотя не должен был")
			for k, v := range handlers.pathTimesCalled {
				assert.Contains(t, tc.expectedPathTimesCalled, k, "Неправильно зароутился запрос, отсутствует ключ")
				assert.Equal(t, tc.expectedPathTimesCalled[k], v, "Неправильно зароутился запрос, не дернулась нужная ручка")
			}
			handlers.pathTimesCalled = map[string]int64{}
		})
	}
}
