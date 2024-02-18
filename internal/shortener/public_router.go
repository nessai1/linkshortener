package shortener

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"go.uber.org/zap"

	"github.com/go-chi/chi"
)

func (application *Application) handleAddURL(writer http.ResponseWriter, request *http.Request) {
	userUUIDCtxValue := request.Context().Value(app.ContextUserUUIDKey)
	if userUUIDCtxValue == nil {
		writer.WriteHeader(http.StatusForbidden)
		application.logger.Error("No user UUID assigned")
		return
	}

	userUUID := userUUIDCtxValue.(app.UserUUID)

	body, err := io.ReadAll(request.Body)
	if err != nil {
		application.logger.Debug("Client sends invalid request", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Failed to read body."))
		return
	}

	if !validateURL(body) {
		application.logger.Debug("Client sends invalid url", zap.String("url", string(body)))
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Invalid pattern of given URI"))
		return
	}

	hash, err := application.createResource(
		request.Context(),
		linkstorage.Link{
			Value:     string(body),
			OwnerUUID: string(userUUID),
		},
	)

	if err != nil {
		if errors.Is(err, linkstorage.ErrURLIntersection) {
			writer.WriteHeader(http.StatusConflict)
			application.logger.Debug(fmt.Sprintf("User insert dublicate url: %s", string(body)))
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
			application.logger.Error(fmt.Sprintf("Cannot create resource for \"%s\"", body), zap.Error(err))
			return
		}
	}

	link := application.buildTokenTail() + hash

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(link))
	application.logger.Debug("Client success add URL", zap.String("URL", link))
}

func (application *Application) handleGetURL(writer http.ResponseWriter, request *http.Request) {
	token := chi.URLParam(request, "token")
	if token == "" {
		application.logger.Debug("Client sends empty request")
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	link, ok := application.storage.Get(request.Context(), token)
	if !ok {
		application.logger.Debug(fmt.Sprintf("Link storage doesn't contain link \"%s\"", link.Value))
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if link.IsDeleted {
		application.logger.Debug(fmt.Sprintf("Client success get resource \"%s\", but it's was deleted", link.Value))
		writer.WriteHeader(http.StatusGone)
		return
	}

	application.logger.Debug(fmt.Sprintf("Client success redirected from \"%s\" to \"%s\"", application.GetAddr()+"/"+token, link.Value))
	writer.Header().Set("Location", link.Value)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (application *Application) handleCheckStorageStatus(writer http.ResponseWriter, request *http.Request) {
	driverIsOk, err := application.storage.Ping()

	if !driverIsOk {
		application.logger.Error("Error while ping storage", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
}

func (application *Application) getPublicRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(app.GetRequestLogMiddleware(application.logger, "PUBLIC"))
	router.Use(app.GetRegisterMiddleware(application.logger))

	router.Post("/", application.handleAddURL)
	router.Get("/{token}", application.handleGetURL)
	router.Get("/ping", application.handleCheckStorageStatus)

	return router
}
