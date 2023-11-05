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

func Run(application Application, envType EnvType) {
	router := chi.NewRouter()

	logger, err := CreateAppLogger(envType)
	if err != nil {
		panic(fmt.Sprintf("Got error while creating application logger: %s", err.Error()))
	}

	defer logger.Sync()

	application.SetLogger(logger)

	router.Use(getZipMiddleware(logger))

	for _, controller := range application.GetControllers() {
		router.Mount(controller.Path, controller.Mux)
	}

	logger.Info("Starting server", zap.String("Server addr", application.GetAddr()))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		application.OnBeforeClose()
		os.Exit(0)
	}()

	if err := http.ListenAndServe(application.GetAddr(), router); err != nil {
		panic(err)
	}
	defer application.OnBeforeClose()
}

type Application interface {
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
