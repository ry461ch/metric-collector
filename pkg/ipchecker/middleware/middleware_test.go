package ipcheckermiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"

	"github.com/ry461ch/metric-collector/pkg/ipchecker"
)

func mockRouter(ipChecker *ipchecker.IPChecker) chi.Router {
	router := chi.NewRouter()
	router.Use(CheckRequesterIP(ipChecker))
	router.Post("/*", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
	})
	return router
}

func TestBase(t *testing.T) {
	ipChecker := ipchecker.New("127.0.0.1/30")
	router := mockRouter(ipChecker)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New()
	resp, _ := client.R().SetHeader("X-Real-IP", "127.0.0.2").Post(srv.URL + "/")
	assert.Equal(t, http.StatusOK, resp.StatusCode(), "Invalid status code")

	resp, _ = client.R().SetHeader("X-Real-IP", "127.0.1.1").Post(srv.URL + "/")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode(), "Invalid status code")
}
