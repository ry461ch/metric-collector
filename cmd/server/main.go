package main

import (
	"net/http"
	"strconv"
	"strings"
)

type Middleware func(http.Handler) http.Handler

type memStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}

var metrics = memStorage{
	counter: make(map[string]int64),
	gauge:   make(map[string]float64)}

func Conveyor(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func middlewareAllowMethodPost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		next.ServeHTTP(res, req)
	})
}

func middlewareValidateRequestData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		args := strings.Split(req.URL.Path, "/")

		if len(args) != 2 {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		next.ServeHTTP(res, req)
	})
}

func gaugeHandler(res http.ResponseWriter, req *http.Request) {
	args := strings.Split(req.URL.Path, "/")
	val, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metrics.gauge[args[0]] = val
	res.WriteHeader(http.StatusOK)
}

func counterHandler(res http.ResponseWriter, req *http.Request) {
	args := strings.Split(req.URL.Path, "/")
	val, err := strconv.ParseInt(args[1], 10, 0)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metrics.counter[args[0]] += val
	res.WriteHeader(http.StatusOK)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle(
		`/update/gauge/`,
		http.StripPrefix(`/update/gauge/`, Conveyor(
			http.HandlerFunc(gaugeHandler),
			middlewareAllowMethodPost,
			middlewareValidateRequestData)))
	mux.Handle(
		`/update/counter/`,
		http.StripPrefix(`/update/counter/`, Conveyor(
			http.HandlerFunc(counterHandler),
			middlewareAllowMethodPost,
			middlewareValidateRequestData)))
	mux.HandleFunc(`/update/`, func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusBadRequest)
	})

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
