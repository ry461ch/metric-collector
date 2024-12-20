package rsamiddleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/ry461ch/metric-collector/pkg/rsa"
)

// Проверка подписи пришедшего запроса
func DecryptResponse(decrypter *rsa.RsaDecrypter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var buf bytes.Buffer
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			reqBody := buf.Bytes()
			reqDecrypted, err := decrypter.Decrypt(reqBody)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(reqDecrypted))
			next.ServeHTTP(w, r)
		})
	}
}
