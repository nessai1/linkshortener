package app

import "github.com/go-chi/chi"

// Controller структура группы обработчиков приложения, закрепленная за конкретным путем
type Controller struct {
	// Path путь по которому работает контроллер
	Path string
	// Mux группа обработчиков
	Mux *chi.Mux
}
