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
			HandlerFunc: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodGet {
					handleGetURL(writer, request)
				} else if request.Method == http.MethodPost {
					handleAddURL(writer, request)
				} else {
					writer.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
		},
	}
}

func (application *Application) GetAddr() string {
	return ":8080"
}

func handleAddURL(w http.ResponseWriter, r *http.Request) {

}

func handleGetURL(w http.ResponseWriter, r *http.Request) {

}
