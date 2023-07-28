package app

import (
	"fmt"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"net/http"
)

func Run(handler ApplicationHandler, envType EnvType) {
	router := chi.NewRouter()

	logger, err := createAppLogger(envType)
	if err != nil {
		panic(fmt.Sprintf("Got error while creating application logger: %s", err.Error()))
	}

	defer logger.Sync()

	handler.SetLogger(logger)
	router.Use(getRequestLogMiddleware(logger))

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

	// SetLogger TODO: выделить только необходимые функции логирования (функции Warn, Info etc.) в отдельный интерфейс.
	// Таким образом уйдет зависимость от конкретной библиотки
	SetLogger(logger *zap.Logger)
}

type EnvType uint8

const (
	Production  EnvType = 0
	Stage       EnvType = 1
	Development EnvType = 2
)
