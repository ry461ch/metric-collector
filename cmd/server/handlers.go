package main

import (
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.Handler

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

func middlewareValidateArgsNum(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		args := strings.Split(req.URL.Path, "/")

		if len(args) != 2 && !(len(args) == 3 && args[2] == "") {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		next.ServeHTTP(res, req)
	})
}
