package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/internal/server/middlewares"
)

func New(mHandlers metricHandlers) chi.Router {
	router := chi.NewRouter()
	router.Use(middlewares.WithLogging)

	router.Route("/update", func(r chi.Router) {
		r.Route("/counter", func(r chi.Router) {
			r.Use(middlewares.ValidatePlainContentType)
			r.Route("/{name:[a-zA-Z0-9-_]+}", func(r chi.Router) {
				r.Post("/{value:[0-9]+}", mHandlers.PostPlainCounterHandler)
				r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
					res.WriteHeader(http.StatusBadRequest)
				})
			})
			r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Route("/gauge", func(r chi.Router) {
			r.Use(middlewares.ValidatePlainContentType)
			r.Route("/{name:[a-zA-Z0-9-_]+}", func(r chi.Router) {
				r.Post("/{value:[0-9]+\\.?[0-9]*}", mHandlers.PostPlainGaugeHandler)
				r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
					res.WriteHeader(http.StatusBadRequest)
				})
			})
			r.Post("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Route("/", func(r chi.Router) {
			r.Use(middlewares.ValidateJSONContentType)
			r.Post("/", mHandlers.PostJSONHandler)
		})
		r.Post("/+", func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusBadRequest)
		})
	})
	router.Route("/value", func(r chi.Router) {
		r.Route("/counter", func(r chi.Router) {
			r.Use(middlewares.ValidatePlainContentType)

			r.Get("/{name:[a-zA-Z0-9-_]+}", mHandlers.GetPlainCounterHandler)
			r.Get("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Route("/gauge", func(r chi.Router) {
			r.Use(middlewares.ValidatePlainContentType)

			r.Get("/{name:[a-zA-Z0-9-_]+}", mHandlers.GetPlainGaugeHandler)
			r.Get("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Route("/", func(r chi.Router) {
			r.Use(middlewares.ValidateJSONContentType)
			r.Post("/", mHandlers.GetJSONHandler)
		})
		r.Get("/+", func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusNotFound)
		})
	})
	router.Route("/", func(r chi.Router) {
		r.Use(middlewares.ValidatePlainContentType)
		router.Get("/", mHandlers.GetPlainAllMetricsHandler)
	})
	return router
}
