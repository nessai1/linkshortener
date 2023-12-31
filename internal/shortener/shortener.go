package shortener

import (
	"errors"
	"fmt"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener/encoder"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"net/http"
	"regexp"

	"go.uber.org/zap"
)

type Config struct {
	ServerAddr    string
	TokenTail     string
	StorageDriver linkstorage.StorageDriver
}

func GetApplication(config *Config) *Application {

	lstorage, err := linkstorage.CreateStorage(config.StorageDriver)
	if err != nil {
		panic(fmt.Sprintf("cannot create storage with driver: %s", err.Error()))
	}

	application := Application{
		config:  config,
		storage: lstorage,
	}

	return &application
}

type Application struct {
	config  *Config
	logger  *zap.Logger
	storage *linkstorage.Storage
}

func (application *Application) OnBeforeClose() {
	application.logger.Info("Closing shorter application...")
	err := application.storage.Save()
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
			Path: "/api/",
		},
	}
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

func (application *Application) createResource(link linkstorage.Link) (string, error) {
	hash, err := encoder.EncodeURL(link.Value)
	if err != nil {
		return "", err
	}

	err = application.storage.Set(hash, link)
	if err != nil && !errors.Is(err, linkstorage.ErrURLIntersection) {
		return "", err
	}

	return hash, err
}

func validateURL(url []byte) bool {
	res, err := regexp.Match(`^https?://[^\s]+$`, url)
	if err != nil {
		return false
	}
	return res
}
