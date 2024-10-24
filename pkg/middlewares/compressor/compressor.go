package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Архиватор запросов
type Compressor struct {
	http.ResponseWriter
	Writer io.Writer
}

// проксирование метода Write в ResponseWiter
func (w Compressor) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Миддлваря архиватора
func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") ||
			(contentType != "" && contentType != "application/json" &&
				!strings.Contains(contentType, "text/html")) {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(Compressor{ResponseWriter: w, Writer: gz}, r)
	})
}
