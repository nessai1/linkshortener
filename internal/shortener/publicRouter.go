package shortener

import (
	"github.com/go-chi/chi"
	"net/http"
)

func (application *Application) getPublicRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/some", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("this is public controller"))
	})

	return router
}
