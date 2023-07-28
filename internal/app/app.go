package app

import (
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
)

func Run(handler ApplicationHandler, envType EnvType) {
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			fmt.Println("Before")
			next.ServeHTTP(writer, request)
			fmt.Println("After")
		})
	})

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

type EnvType uint8

const (
	Production  EnvType = 0
	Stage       EnvType = 1
	Development EnvType = 2
)
