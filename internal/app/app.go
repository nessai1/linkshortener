package app

import (
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// ApplicationInfo информация о приложении, указываямая при сборке линтером
type ApplicationInfo struct {
	// BuildVersion Версия билда приложения
	BuildVersion string
	// BuildDate Дата билда приложения
	BuildDate string
	// BuildCommit последний коммит билда приложения
	BuildCommit string
}

// Run запускает реализацию Application с режимом работы EnvType
func Run(application Application, envType EnvType, info ApplicationInfo, useSecure bool) {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", info.BuildVersion, info.BuildDate, info.BuildCommit)

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

	if envType == Development {
		router.Mount("/debug", middleware.Profiler())
	}

	logger.Info("Starting server", zap.String("Server addr", application.GetAddr()))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		application.OnBeforeClose()
		os.Exit(0)
	}()

	if useSecure {
		server := buildSecureServer(application.GetAddr(), router)
		err = server.ListenAndServeTLS("", "")
	} else {
		err = http.ListenAndServe(application.GetAddr(), router)
	}

	if err != nil {
		log.Fatalf("Error while start listening server: %s", err.Error())
	}

	defer application.OnBeforeClose()
}

func buildSecureServer(addr string, mux http.Handler) *http.Server {
	// конструируем менеджер TLS-сертификатов
	manager := &autocert.Manager{
		// директория для хранения сертификатов
		Cache: autocert.DirCache("cache-dir"),
		// функция, принимающая Terms of Service издателя сертификатов
		Prompt: autocert.AcceptTOS,
		// перечень доменов, для которых будут поддерживаться сертификаты
		HostPolicy: autocert.HostWhitelist(addr),
	}

	// конструируем сервер с поддержкой TLS
	return &http.Server{
		Addr:    addr,
		Handler: mux,
		// для TLS-конфигурации используем менеджер сертификатов
		TLSConfig: manager.TLSConfig(),
	}
}

// Application интерфейс описывающий методы, на которые уделяется ответсвенность при работе фасада веб-приложения.
type Application interface {
	// GetAddr возвращает адрес по которому будет запущено приложение
	GetAddr() string

	// SetLogger вызывается фасадом, который передает логгер приложения в реализацию
	SetLogger(logger *zap.Logger)

	// OnBeforeClose метод вызывается когда приложение получает сигнал о завершении работы
	OnBeforeClose()

	// GetControllers возвращает список контроллеров, которые будут смонтированы в роутер приложения
	GetControllers() []Controller
}

// EnvType идентификатор режима работы приложения
type EnvType uint8

// Список всех доступных режимов работы приложения
const (
	// Production режим работы в продакшн среде, в лог идет только ошибки
	Production EnvType = iota

	// Stage режим работы в stage среде, в лог идет информационные логи, помимо ошибок
	Stage

	// Development режим работы для разработки, в лог идут дебаг-логи, добавляются обработчики для профилировщика pprof
	Development
)
