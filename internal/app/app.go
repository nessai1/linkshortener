package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
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

	var server http.Server
	var start func(server *http.Server) error

	if useSecure {
		server = buildSecureServer(application.GetAddr(), router)
		start = func(server *http.Server) error {
			return server.ListenAndServeTLS("", "")
		}
	} else {
		server = http.Server{
			Addr:    application.GetAddr(),
			Handler: router,
		}
		start = func(server *http.Server) error {
			return server.ListenAndServe()
		}
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	var gRPCServer *grpc.Server
	if application.GetGRPCAddr() != "" {
		listen, err := net.Listen("tcp", application.GetGRPCAddr())
		if err != nil {
			logger.Fatal("Error while start gRPC listener", zap.Error(err))
			return
		}

		interceptors := application.GetGRPCInterceptors()
		if interceptors != nil {
			gRPCServer = grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
		} else {
			gRPCServer = grpc.NewServer()
		}

		err = application.RegisterGRPCService(gRPCServer)
		if err != nil {
			logger.Fatal("Error while register gRPC service", zap.Error(err))
			return
		}

		go func() {
			logger.Info("Starting gRPC server", zap.String("Server addr", application.GetGRPCAddr()))
			if err := gRPCServer.Serve(listen); err != nil {
				logger.Fatal("Cannot start gRPC server", zap.Error(err))
				sigint <- os.Interrupt
			}
		}()
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		<-sigint
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Error("Error HTTP server shutdown listener", zap.Error(err))
		}

		if gRPCServer != nil {
			gRPCServer.GracefulStop()
		}
		close(idleConnsClosed)
	}()

	if err := start(&server); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("Error while start listening server", zap.Error(err))
		return
	}

	<-idleConnsClosed
	application.OnBeforeClose()
	logger.Info("Server graceful shutdown")
}

func buildSecureServer(addr string, mux http.Handler) http.Server {
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
	return http.Server{
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

	// GetGRPCAddr возвращает адрес по которому будет запущен gRPC сервер. Если пуста строчка - сервер не запустится
	GetGRPCAddr() string

	// GetGRPCInterceptors возвращает список перехватчиков gRPC сервера
	GetGRPCInterceptors() []grpc.UnaryServerInterceptor

	// RegisterGRPCService регистрирует gRPC сервер, исходящий из кода старта приложения
	RegisterGRPCService(server *grpc.Server) error
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
