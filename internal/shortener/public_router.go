package shortener

import (
	"errors"
	"fmt"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"io"
	"net/http"

	"github.com/go-chi/chi"
)

func (application *Application) handleAddURL(writer http.ResponseWriter, request *http.Request) {
	UserUUID, err := app.Authorize(writer, request)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		application.logger.Error(fmt.Sprintf("Cannot authorize user: %s", err.Error()))
		return
	}
	application.logger.Info(fmt.Sprintf("User auth: %s", UserUUID))

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

	hash, err := application.createResource(linkstorage.Link{
		Value:     string(body),
		OwnerUUID: &UserUUID,
	})

	if err != nil {
		if errors.Is(err, linkstorage.ErrURLIntersection) {
			writer.WriteHeader(http.StatusConflict)
			application.logger.Debug(fmt.Sprintf("User insert dublicate url: %s", string(body)))
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
			application.logger.Debug(fmt.Sprintf("Cannot create resource for \"%s\". (%s)", body, err.Error()))
			application.logger.Error(fmt.Sprintf("Error while creating resource '%s'\n", body))
			return
		}
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

	link, ok := application.storage.Get(token)
	if !ok {
		application.logger.Debug(fmt.Sprintf("Link storage doesn't contain link \"%s\"", link.Value))
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if link.IsDeleted {
		application.logger.Info(fmt.Sprintf("Client success get resource \"%s\", but it's was deleted", link.Value))
		writer.WriteHeader(http.StatusGone)
		return
	}

	application.logger.Info(fmt.Sprintf("Client success redirected from \"%s\" to \"%s\"", application.GetAddr()+"/"+token, link.Value))
	writer.Header().Set("Location", link.Value)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (application *Application) handleCheckStorageStatus(writer http.ResponseWriter, request *http.Request) {
	driverIsOk, err := application.storage.Ping()

	if !driverIsOk {
		application.logger.Info("Error while ping storage: " + err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
}

func (application *Application) getPublicRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(app.GetRequestLogMiddleware(application.logger, "PUBLIC"))

	router.Post("/", application.handleAddURL)
	router.Get("/{token}", application.handleGetURL)
	router.Get("/ping", application.handleCheckStorageStatus)

	return router
}
