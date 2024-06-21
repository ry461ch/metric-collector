package middlewares

import (
	"net/http"
)

func ValidateJsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		if contentType != "application/json" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		next.ServeHTTP(res, req)
		res.Header().Set("Content-Type", "application/json")
	})
}

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
