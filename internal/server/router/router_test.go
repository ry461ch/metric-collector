package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"
)

type MockHandlers struct {
	pathTimesCalled map[string]int64
}

func NewMockHandlers() MockHandlers {
	return MockHandlers{pathTimesCalled: map[string]int64{}}
}

func (m *MockHandlers) PostGaugeHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["postGauge"] += 1
}

func (m *MockHandlers) PostCounterHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["postCounter"] += 1
}

func (m *MockHandlers) GetGaugeHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["getGauge"] += 1
}

func (m *MockHandlers) GetCounterHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["getCounter"] += 1
}

func (m *MockHandlers) GetAllMetricsHandler(res http.ResponseWriter, req *http.Request) {
	m.pathTimesCalled["getAll"] += 1
}

func TestRouter(t *testing.T) {
	defaultPostGaugeRequest := "/update/gauge/some_metric/10.0"
	defaultPostCounterRequest := "/update/counter/some_metric/10"
	defaultGetGaugeRequest := "/value/gauge/some_metric"
	defaultGetCounterRequest := "/value/counter/some_metric"

	testCases := []struct {
		testName     string
		method       string
		requestPath  string
		expectedCode int
		expectedPathTimesCalled map[string]int64
	}{
		{
			testName: "invalid method",
			method: http.MethodDelete,
			requestPath: defaultPostCounterRequest,
			expectedCode: http.StatusMethodNotAllowed,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "ok for post gauge",
			method: http.MethodPost,
			requestPath: defaultPostGaugeRequest,
			expectedCode: http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"postGauge": 1},
		},
		{
			testName: "ok for post counter",
			method: http.MethodPost,
			requestPath: defaultPostCounterRequest,
			expectedCode: http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"postCounter": 1},
		},
		{
			testName: "ok for get counter",
			method: http.MethodGet,
			requestPath: defaultGetCounterRequest,
			expectedCode: http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"getCounter": 1},
		},
		{
			testName: "ok for get gauge",
			method: http.MethodGet,
			requestPath: defaultGetGaugeRequest,
			expectedCode: http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"getGauge": 1},
		},
		{
			testName: "ok for get all",
			method: http.MethodGet,
			requestPath: "/",
			expectedCode: http.StatusOK,
			expectedPathTimesCalled: map[string]int64{"getAll": 1},
		},
		{
			testName: "no metric name and value for post counter",
			method: http.MethodPost,
			requestPath: "/update/counter/",
			expectedCode: http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "no metric name and value for post gauge",
			method: http.MethodPost,
			requestPath: "/update/gauge/",
			expectedCode: http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "invalid metric type for post",
			method: http.MethodPost,
			requestPath: "/update/invalid_metric_type",
			expectedCode: http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "no metric name for post counter",
			method: http.MethodPost,
			requestPath: "/update/counter/10",
			expectedCode: http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "no metric name for post gauge",
			method: http.MethodPost,
			requestPath: "/update/gauge/10",
			expectedCode: http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "invalid metric value for post counter",
			method: http.MethodPost,
			requestPath: "/update/counter/test/10.0",
			expectedCode: http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "invalid metric value for post gauge",
			method: http.MethodPost,
			requestPath: "/update/gauge/test/str",
			expectedCode: http.StatusBadRequest,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "no metric name for get counter",
			method: http.MethodGet,
			requestPath: "/value/counter/",
			expectedCode: http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "no metric name for get gauge",
			method: http.MethodGet,
			requestPath: "/value/gauge/",
			expectedCode: http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "no metric type for get",
			method: http.MethodGet,
			requestPath: "/value/invalid/",
			expectedCode: http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
		{
			testName: "bad path for get",
			method: http.MethodGet,
			requestPath: "/invalid",
			expectedCode: http.StatusNotFound,
			expectedPathTimesCalled: map[string]int64{},
		},
	}

	client := resty.New()

	handlers := NewMockHandlers()
	router := New(&handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			resp, err := client.R().Execute(tc.method, srv.URL+tc.requestPath)
			assert.Nil(t, err, "Сервер вернул 500")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, len(tc.expectedPathTimesCalled), len(handlers.pathTimesCalled), "Запрос прошел до сервиса, хотя не должен был")
			for k, v := range(handlers.pathTimesCalled) {
				assert.Contains(t, tc.expectedPathTimesCalled, k, "Неправильно зароутился запрос, отсутствует ключ")
				assert.Equal(t, tc.expectedPathTimesCalled[k], v, "Неправильно зароутился запрос, не дернулась нужная ручка")
			}
			handlers.pathTimesCalled = map[string]int64{}
		})
	}
}
