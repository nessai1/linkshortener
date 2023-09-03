package shortener

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/nessai1/linkshortener/internal/app"
	encoder "github.com/nessai1/linkshortener/internal/shortener/encoder"
	"github.com/nessai1/linkshortener/internal/storage"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"regexp"
)

type Config struct {
	ServerAddr  string
	TokenTail   string
	StoragePath string
}

func GetApplication(config *Config, innerStorage *storage.KeyValueStorage) *Application {
	application := Application{
		config:  config,
		storage: innerStorage,
	}

	return &application
}

type Application struct {
	config  *Config
	logger  *zap.Logger
	storage *storage.KeyValueStorage
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
		{
			URL: "/api",
			Group: []app.Endpoint{
				{
					URL:         "/shorten",
					Method:      http.MethodPost,
					HandlerFunc: application.apiHandleAddURL,
				},
			},
		},
	}
}

func (application *Application) OnBeforeClose() {
	application.logger.Info("Closing shorter application...")
	err := application.storage.Close()
	if err != nil {
		application.logger.Error("Error while closing application storage")
	} else {
		application.logger.Info("Application storage is closed successful")
	}
}

func (application *Application) GetAddr() string {
	if configAddr := application.config.ServerAddr; configAddr != "" {
		return configAddr
	}

	return "localhost:8080"
}

func (application *Application) SetLogger(logger *zap.Logger) {
	application.logger = logger
}

func (application *Application) GetControllers() []app.Controller {
	return []app.Controller{
		{
			Mux:  application.getPublicRouter(),
			Path: "/",
		},
		{
			Mux:  application.getAPIRouter(),
			Path: "/api",
		},
	}
}

func (application *Application) handleAddURL(writer http.ResponseWriter, request *http.Request) {

	body, err := io.ReadAll(request.Body)
	if err != nil {
		application.logger.Debug(fmt.Sprintf("Client sends invalid request. (%s)", err.Error()))
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Failed to read body."))
		return
	}

	if !validateURL(body) {
		application.logger.Debug(fmt.Sprintf("Client sends invalid url: %s", body))
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Invalid pattern of given URI"))
		return
	}

	hash, err := application.createResource(string(body))
	if err != nil {
		application.logger.Debug(fmt.Sprintf("Cannot create resource for \"%s\". (%s)", body, err.Error()))
		log.Printf("Error while creating resource '%s'\n", body)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Error while creating resource!"))
		return
	}

	link := application.buildTokenTail(request) + hash

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(link))
	application.logger.Info(fmt.Sprintf("Client success add URL \"%s\"", link))
}

func (application *Application) handleGetURL(writer http.ResponseWriter, request *http.Request) {
	token := chi.URLParam(request, "token")
	if token == "" {
		application.logger.Debug("Client sends empty request")
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	uri, ok := application.storage.Get(token)
	if !ok {
		application.logger.Debug(fmt.Sprintf("Link storage doesn't contain link \"%s\"", uri))
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	application.logger.Info(fmt.Sprintf("Client success redirected from \"%s\" to \"%s\"", application.GetAddr()+"/"+token, uri))
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

	application.storage.Set(hash, url)

	return hash, nil
}
