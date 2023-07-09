package shortener

import (
	"github.com/nessai1/linkshortener/internal/app"
	encoder "github.com/nessai1/linkshortener/internal/shortener/decoder"
	"io"
	"log"
	"net/http"
	"regexp"
)

func GetApplication() *Application {
	application := Application{
		links: map[string]string{},
	}

	return &application
}

type Application struct {
	links map[string]string
}

func (application *Application) GetEndpoints() []app.Endpoint {
	return []app.Endpoint{
		{
			Url: "/",
			HandlerFunc: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodGet {
					application.handleGetURL(writer, request)
				} else if request.Method == http.MethodPost {
					application.handleAddURL(writer, request)
				} else {
					writer.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
		},
	}
}

func (application *Application) GetAddr() string {
	return "localhost:8080"
}

func (application *Application) handleAddURL(writer http.ResponseWriter, request *http.Request) {

	body, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Failed to read body."))
		return
	}

	if !validateURL(body) {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Invalid pattern of given URI"))
		return
	}

	hash, err := application.createResource(string(body))
	if err != nil {
		log.Printf("Error while creating resource '%s'\n", body)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Error while creating resource!"))
		return
	}

	scheme := "http://"
	if request.TLS != nil {
		scheme = "https://"
	}
	link := scheme + application.GetAddr() + "/" + hash

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(link))
	return
}

func (application *Application) handleGetURL(writer http.ResponseWriter, request *http.Request) {
	token := request.URL.Path[1:]
	if token == "" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	URI, ok := application.links[token]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.Header().Set("Location", URI)
	writer.WriteHeader(http.StatusTemporaryRedirect)
	return
}

func validateURL(url []byte) bool {
	res, err := regexp.Match(`^https?://[^\s]+$`, url)
	if err != nil {
		return false
	}

	return res
}

func (application *Application) createResource(url string) (string, error) {
	hash, err := encoder.EncodeURL(url)
	if err != nil {
		return "", err
	}

	application.links[hash] = url

	return hash, nil
}
