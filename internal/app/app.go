package app

import (
	"github.com/go-chi/chi"
	"net/http"
)

func Run(handler ApplicationHandler) {
	router := chi.NewRouter()
	fillRouter(router, handler.GetEndpoints())
	if err := http.ListenAndServe(handler.GetAddr(), router); err != nil {
		panic(err)
	}
}

func fillRouter(router chi.Router, endpoints []Endpoint) {
	for _, endpoint := range endpoints {
		method := endpoint.Method
		if method == "" {
			method = http.MethodGet
		}

		router.MethodFunc(method, endpoint.URL, endpoint.HandlerFunc)
	}
}

type Endpoint struct {
	URL         string
	Method      string
	HandlerFunc func(http.ResponseWriter, *http.Request)
}

type ApplicationHandler interface {
	GetEndpoints() []Endpoint
	GetAddr() string
}
