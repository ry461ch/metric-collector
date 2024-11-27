package ipcheckermiddleware

import (
	"net"
	"net/http"

	"github.com/ry461ch/metric-collector/pkg/ipchecker"
)

// Проверка подписи пришедшего запроса
func CheckRequesterIP(ipChecker *ipchecker.IPChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ip := req.Header.Get("X-Real-IP")
			realIP := net.ParseIP(ip)
			if realIP == nil || !ipChecker.Contains(&realIP) {
				res.WriteHeader(http.StatusBadRequest)
				return
			}

			next.ServeHTTP(res, req)
		})
	}
}
