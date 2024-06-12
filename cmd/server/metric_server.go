package main

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/internal/storage"
)

type MetricUpdateServer struct {
	mStorage storage.Storage
}

func (server *MetricUpdateServer) postGaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseFloat(chi.URLParam(req, "value"), 64)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	server.mStorage.UpdateGaugeValue(metricName, metricVal)
	res.WriteHeader(http.StatusOK)
}

func (server *MetricUpdateServer) postCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseInt(chi.URLParam(req, "value"), 10, 0)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	server.mStorage.UpdateCounterValue(metricName, metricVal)
	res.WriteHeader(http.StatusOK)
}

func (server *MetricUpdateServer) getCounterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	val, ok := server.mStorage.GetCounterValue(metricName)

	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	io.WriteString(res, strconv.FormatInt(val, 10))
}

func (server *MetricUpdateServer) getGaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	val, ok := server.mStorage.GetGaugeValue(metricName)

	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	io.WriteString(res, strconv.FormatFloat(val, 'f', -1, 64))
}

func (server *MetricUpdateServer) getAllMetricsHandler(res http.ResponseWriter, req *http.Request) {
	gaugeMetrics := server.mStorage.GetGaugeValues()
	counterMetrics := server.mStorage.GetCounterValues()

	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	for name, val := range gaugeMetrics {
		io.WriteString(res, name+" : "+strconv.FormatFloat(val, 'f', -1, 64)+"\n")
	}
	for name, val := range counterMetrics {
		io.WriteString(res, name+" : "+strconv.FormatInt(val, 10)+"\n")
	}
}

func (server *MetricUpdateServer) MakeRouter() chi.Router {
	router := chi.NewRouter()
	router.Use(middlewareValidateContentType)

	router.Route("/update", func(r chi.Router) {
		r.Route("/counter", func(r chi.Router) {
			r.Route("/{name:[a-zA-Z0-9-_]+}", func(r chi.Router) {
				r.Post("/{value:[0-9]+}", server.postCounterHandler)
				r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
					res.WriteHeader(http.StatusBadRequest)
				})
			})
			r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Route("/gauge", func(r chi.Router) {
			r.Route("/{name:[a-zA-Z0-9-_]+}", func(r chi.Router) {
				r.Post("/{value:[0-9]+\\.?[0-9]*}", server.postGaugeHandler)
				r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
					res.WriteHeader(http.StatusBadRequest)
				})
			})
			r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusBadRequest)
		})
	})
	router.Route("/value", func(r chi.Router) {
		r.Route("/counter", func(r chi.Router) {
			r.Get("/{name:[a-zA-Z0-9-_]+}", server.getCounterHandler)
			r.Get("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Route("/gauge", func(r chi.Router) {
			r.Get("/{name:[a-zA-Z0-9-_]+}", server.getGaugeHandler)
			r.Get("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Get("/*", func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusNotFound)
		})
	})
	router.Get("/", server.getAllMetricsHandler)
	return router
}
