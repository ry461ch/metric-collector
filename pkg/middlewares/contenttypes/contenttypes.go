package contenttypes

import (
	"net/http"
)

// Миддлваря проверки запроса на content-type==application/json
func ValidateJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		if contentType != "application/json" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		res.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(res, req)
	})
}

// Миддлваря проверки запроса на content-type==text/plain
func ValidatePlainContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		if contentType != "" && contentType != "text/plain" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		next.ServeHTTP(res, req)
	})
}
