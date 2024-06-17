package middlewares

import (
	"net/http"
)

func ValidateContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		if contentType != "" && contentType != "text/plain" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		next.ServeHTTP(res, req)
	})
}
