package app

import (
	"log"
	"net/http"
)

func Run(handler ApplicationHandler) {
	mux := http.NewServeMux()
	fillMux(mux, handler.GetEndpoints())
	log.Fatalln(http.ListenAndServe(handler.GetAddr(), mux))
}

func fillMux(mux *http.ServeMux, endpoints []Endpoint) {
	for _, endpoint := range endpoints {
		mux.HandleFunc(endpoint.Url, endpoint.HandlerFunc)
	}
}

type Endpoint struct {
	Url         string
	HandlerFunc func(http.ResponseWriter, *http.Request)
}

type ApplicationHandler interface {
	GetEndpoints() []Endpoint
	GetAddr() string
}
