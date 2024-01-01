package app_test

import (
	"github.com/go-chi/chi"
	"github.com/nessai1/linkshortener/internal/app"
	"go.uber.org/zap"
	"net/http"
)

type MyBeautifulApp struct {
	logger *zap.Logger
}

func (m *MyBeautifulApp) GetAddr() string {
	// Возвращаем адрес, по которому будет запущено приложение
	return ":1337"
}

func (m *MyBeautifulApp) SetLogger(logger *zap.Logger) {
	// Получаем логгер из внешнего источника
	m.logger = logger
}

func (m *MyBeautifulApp) OnBeforeClose() {
	// Пишем в лог перед остановкой приложения
	m.logger.Info("bye!")
}

func (m *MyBeautifulApp) GetControllers() []app.Controller {

	// Создаем тестовую группу обработчиквов
	mux := chi.NewMux()
	mux.HandleFunc("/ping", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Write([]byte("pong"))
	})

	// Создаем новый контроллер, привязывая группу обработчиков к подпути "/testpath"
	controller := app.Controller{
		Path: "testpath",
		Mux:  mux,
	}

	// Создаем и возвращаем список контроллеров
	controllers := make([]app.Controller, 1)
	controllers[0] = controller

	return controllers
}

func ExampleRun() {
	// Определяем реализацию приложения, которое будем запускать
	myApp := MyBeautifulApp{}

	// Запускаем приложение в режиме работы для разработки
	app.Run(&myApp, app.Development)
}
