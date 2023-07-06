package shortener

import (
	"github.com/nessai1/linkshortener/internal/app"
	"net/http"
)

func GetApplication() *Application {
	application := Application{}
	return &application
}

type Application struct{}

func (application *Application) GetEndpoints() []app.Endpoint {
	return []app.Endpoint{
		{
			Url: "/",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Hello world!"))
			},
		},
	}
}

func (application *Application) GetAddr() string {
	return ":8080"
}
