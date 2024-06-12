package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/internal/storage"
)

type MetricUpdateServer struct {
	mStorage storage.Storage
}

func (server *MetricUpdateServer) gaugeHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseFloat(chi.URLParam(req, "value"), 64)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	server.mStorage.UpdateGaugeValue(metricName, metricVal)
	res.WriteHeader(http.StatusOK)
}

func (server *MetricUpdateServer) counterHandler(res http.ResponseWriter, req *http.Request) {
	metricName := chi.URLParam(req, "name")
	metricVal, err := strconv.ParseInt(chi.URLParam(req, "value"), 10, 0)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	server.mStorage.UpdateCounterValue(metricName, metricVal)
	res.WriteHeader(http.StatusOK)
}


func (server *MetricUpdateServer) MakeRouter() chi.Router {
	router := chi.NewRouter()
	router.Use(middlewareAllowMethodPost, middlewareValidateContentType)

	router.Route("/update", func(r chi.Router) {
		r.Route("/counter", func(r chi.Router) {
			r.Route("/{name:[a-zA-Z-_]+}", func(r chi.Router) {
				r.Post("/{value:[0-9]+}", server.counterHandler)
				r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
					res.WriteHeader(http.StatusBadRequest)
				})
			})
			r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Route("/gauge", func(r chi.Router) {
			r.Route("/{name:[a-zA-Z-_]+}", func(r chi.Router) {
				r.Post("/{value:[0-9]+\\.?[0-9]*}", server.gaugeHandler)
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
	return router
}
