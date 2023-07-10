package app

import (
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

func Run(handler ApplicationHandler) {
	router := chi.NewRouter()
	fillRouter(router, handler.GetEndpoints())
	log.Fatalln(http.ListenAndServe(handler.GetAddr(), router))
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
