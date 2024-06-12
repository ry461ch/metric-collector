package main

import (
	"net/http"

	"github.com/ry461ch/metric-collector/internal/storage"
)

func main() {
	mux := http.NewServeMux()
	server := MetricUpdateServer{mStorage: &storage.MetricStorage{}}
	mux.HandleFunc(`/update/`, server.UpdateMetricHandler)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
