package shortener

import (
	"net/http"
	"regexp"

	"github.com/nessai1/linkshortener/internal/app"
	encoder "github.com/nessai1/linkshortener/internal/shortener/encoder"
	"github.com/nessai1/linkshortener/internal/storage"
	"go.uber.org/zap"
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
