package app

import (
	"fmt"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(handler ApplicationHandler, envType EnvType) {
	router := chi.NewRouter()

	logger, err := CreateAppLogger(envType)
	if err != nil {
		panic(fmt.Sprintf("Got error while creating application logger: %s", err.Error()))
	}

	defer logger.Sync()

	handler.SetLogger(logger)
	router.Use(getRequestLogMiddleware(logger))
	router.Use(getZipMiddleware(logger))

	fillRouter(router, handler.GetEndpoints(), "")
	logger.Info(fmt.Sprintf("staring server on addr: %s", handler.GetAddr()))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		handler.OnBeforeClose()
		os.Exit(1)
	}()

	if err := http.ListenAndServe(handler.GetAddr(), router); err != nil {
		panic(err)
	}

	defer handler.OnBeforeClose()
}

func fillRouter(router chi.Router, endpoints []Endpoint, tail string) {
	for _, endpoint := range endpoints {
		if len(endpoint.Group) != 0 {
			fillRouter(router, endpoint.Group, endpoint.URL)
			continue
		}

		method := endpoint.Method
		if method == "" {
			method = http.MethodGet
		}

		url := tail + endpoint.URL

		router.MethodFunc(method, url, endpoint.HandlerFunc)
	}
}

type Endpoint struct {
	URL         string
	Method      string
	HandlerFunc func(http.ResponseWriter, *http.Request)
	Group       []Endpoint
}

type ApplicationHandler interface {
	GetEndpoints() []Endpoint
	GetAddr() string

	// SetLogger TODO: выделить только необходимые функции логирования (функции Warn, Info etc.) в отдельный интерфейс.
	// Таким образом уйдет зависимость от конкретной библиотки
	SetLogger(logger *zap.Logger)

	OnBeforeClose()
}

type EnvType uint8

const (
	Production  EnvType = 0
	Stage       EnvType = 1
	Development EnvType = 2
)
