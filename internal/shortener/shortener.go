package shortener

import (
	"github.com/go-chi/chi"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener/config"
	encoder "github.com/nessai1/linkshortener/internal/shortener/encoder"
	"io"
	"log"
	"net/http"
	"regexp"
)

func GetApplication() *Application {
	application := Application{
		links:  map[string]string{},
		config: config.GetConfig(),
	}

	return &application
}

type Application struct {
	links  map[string]string
	config *config.Config
}

func (application *Application) GetEndpoints() []app.Endpoint {
	return []app.Endpoint{
		{
			URL:         "/{token}",
			Method:      http.MethodGet,
			HandlerFunc: application.handleGetURL,
		},
		{
			URL:         "/",
			Method:      http.MethodPost,
			HandlerFunc: application.handleAddURL,
		},
	}
}

func (application *Application) GetAddr() string {
	if configAddr := config.GetConfig().ServerAddr; configAddr != "" {
		return configAddr
	}

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

	link := application.buildTokenTail(request) + hash

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(link))
}

func (application *Application) handleGetURL(writer http.ResponseWriter, request *http.Request) {
	token := chi.URLParam(request, "token")
	if token == "" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	uri, ok := application.links[token]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.Header().Set("Location", uri)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}

func validateURL(url []byte) bool {
	res, err := regexp.Match(`^https?://[^\s]+$`, url)
	if err != nil {
		return false
	}

	return res
}

func (application *Application) buildTokenTail(request *http.Request) string {
	if configTail := application.config.TokenTail; configTail != "" {
		return configTail + "/"
	}

	scheme := "http://"
	if request.TLS != nil {
		scheme = "https://"
	}
	return scheme + application.GetAddr() + "/"
}

func (application *Application) createResource(url string) (string, error) {
	hash, err := encoder.EncodeURL(url)
	if err != nil {
		return "", err
	}

	application.links[hash] = url

	return hash, nil
}
