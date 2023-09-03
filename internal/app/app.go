package app

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func Run(application Application, envType EnvType) {
	router := chi.NewRouter()

	logger, err := CreateAppLogger(envType)
	if err != nil {
		panic(fmt.Sprintf("Got error while creating application logger: %s", err.Error()))
	}

	defer logger.Sync()

	application.SetLogger(logger)

	for _, controller := range application.GetControllers() {
		router.Mount(controller.Path, controller.Mux)
	}

	//router.Use(getRequestLogMiddleware(logger))
	//router.Use(getZipMiddleware(logger))
	//
	//fillRouter(router, application.GetEndpoints(), "")
	logger.Info(fmt.Sprintf("staring server on addr: %s", application.GetAddr()))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		application.OnBeforeClose()
		os.Exit(1)
	}()

	if err := http.ListenAndServe(application.GetAddr(), router); err != nil {
		panic(err)
	}

	defer application.OnBeforeClose()
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

type Application interface {
	GetEndpoints() []Endpoint
	GetAddr() string

	// SetLogger TODO: выделить только необходимые функции логирования (функции Warn, Info etc.) в отдельный интерфейс.
	// Таким образом уйдет зависимость от конкретной библиотки
	SetLogger(logger *zap.Logger)

	OnBeforeClose()

	GetControllers() []Controller
}

type EnvType uint8

const (
	Production  EnvType = 0
	Stage       EnvType = 1
	Development EnvType = 2
)
