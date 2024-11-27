// Module router
package router

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/metric-collector/pkg/encrypt"
	encryptmiddleware "github.com/ry461ch/metric-collector/pkg/encrypt/middleware"
	"github.com/ry461ch/metric-collector/pkg/ipchecker"
	ipcheckermiddleware "github.com/ry461ch/metric-collector/pkg/ipchecker/middleware"
	requestlogger "github.com/ry461ch/metric-collector/pkg/logging/middleware"
	"github.com/ry461ch/metric-collector/pkg/middlewares/compressor"
	"github.com/ry461ch/metric-collector/pkg/middlewares/contenttypes"
	"github.com/ry461ch/metric-collector/pkg/rsa"
	rsamiddleware "github.com/ry461ch/metric-collector/pkg/rsa/middleware"
)

// Router initialization
func New(mHandlers metricHandlers, encrypter *encrypt.Encrypter, rsaDecrypter *rsa.RsaDecrypter, ipChecker *ipchecker.IPChecker) chi.Router {
	r := chi.NewRouter()
	r.Use(
		requestlogger.WithLogging,
		ipcheckermiddleware.CheckRequesterIP(ipChecker),
		compressor.GzipHandle,
		encryptmiddleware.CheckRequestAndEncryptResponse(encrypter),
	)
	if rsaDecrypter != nil {
		r.Use(rsamiddleware.DecryptResponse(rsaDecrypter))
	}

	r.Route("/updates/", func(r chi.Router) {
		r.Use(contenttypes.ValidateJSONContentType)
		r.Post("/", mHandlers.PostMetricsHandler)
	})
	r.Route("/update/", func(r chi.Router) {
		r.Route("/counter/", func(r chi.Router) {
			r.Use(contenttypes.ValidatePlainContentType)
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
		r.Route("/gauge/", func(r chi.Router) {
			r.Use(contenttypes.ValidatePlainContentType)
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
			r.Use(contenttypes.ValidateJSONContentType)
			r.Post("/", mHandlers.PostJSONHandler)
		})
		r.Post("/+", func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusBadRequest)
		})
	})
	r.Route("/value/", func(r chi.Router) {
		r.Route("/counter/", func(r chi.Router) {
			r.Use(contenttypes.ValidatePlainContentType)

			r.Get("/{name:[a-zA-Z0-9-_]+}", mHandlers.GetPlainCounterHandler)
			r.Get("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Route("/gauge/", func(r chi.Router) {
			r.Use(contenttypes.ValidatePlainContentType)

			r.Get("/{name:[a-zA-Z0-9-_]+}", mHandlers.GetPlainGaugeHandler)
			r.Get("/*", func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusNotFound)
			})
		})
		r.Route("/", func(r chi.Router) {
			r.Use(contenttypes.ValidateJSONContentType)
			r.Post("/", mHandlers.GetJSONHandler)
		})
		r.Get("/+", func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusNotFound)
		})
	})
	r.Get("/ping", mHandlers.Ping)
	r.Route("/", func(r chi.Router) {
		r.Use(contenttypes.ValidatePlainContentType)
		r.Get("/", mHandlers.GetPlainAllMetricsHandler)
	})
	r.Route("/debug/pprof/", func(r chi.Router) {
		r.Get("/", pprof.Index)
		r.Get("/cmdline", pprof.Cmdline)
		r.Get("/profile", pprof.Profile)
		r.Get("/symbol", pprof.Symbol)
		r.Get("/trace", pprof.Trace)
		r.Handle("/allocs", pprof.Handler("allocs"))
		r.Handle("/goroutine", pprof.Handler("goroutine"))
		r.Handle("/threadcreate", pprof.Handler("threadcreate"))
		r.Handle("/mutex", pprof.Handler("mutex"))
		r.Handle("/heap", pprof.Handler("heap"))
		r.Handle("/block", pprof.Handler("block"))
	})
	return r
}
