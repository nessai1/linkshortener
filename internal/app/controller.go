package app

import "github.com/go-chi/chi"

type Controller struct {
	Path string
	Mux  *chi.Mux
}
