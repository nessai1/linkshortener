package shortener

import (
	"context"
	"errors"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener/encoder"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"net/http"
	"regexp"

	"go.uber.org/zap"
)

type Config struct {
	ServerAddr  string
	TokenTail   string
	LinkStorage linkstorage.LinkStorage
}

func GetApplication(config *Config) *Application {
	application := Application{
		config:  config,
		storage: config.LinkStorage,
	}

	return &application
}

type Application struct {
	config  *Config
	logger  *zap.Logger
	storage linkstorage.LinkStorage
}

func (application *Application) OnBeforeClose() {
	application.logger.Info("Closing shorter application...")
	err := application.storage.BeforeShutdown()
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

func (application *Application) createResource(ctx context.Context, link linkstorage.Link) (string, error) {
	hash, err := encoder.EncodeURL(link.Value)
	if err != nil {
		return "", err
	}

	err = application.storage.Set(ctx, hash, link)
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
