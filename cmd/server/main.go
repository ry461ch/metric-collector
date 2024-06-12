package main

import (
	"net/http"

	"github.com/ry461ch/metric-collector/internal/storage"
)

func main() {
	server := MetricUpdateServer{mStorage: &storage.MetricStorage{}}
	router := server.MakeRouter()
	err := http.ListenAndServe(`:8080`, router)
	if err != nil {
		panic(err)
	}
}
