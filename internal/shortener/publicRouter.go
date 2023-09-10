package shortener

import (
	"fmt"
	"github.com/nessai1/linkshortener/internal/app"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

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

func (application *Application) handleCheckDBStatus(writer http.ResponseWriter, request *http.Request) {
	driverIsOk := application.SQLDriver.Ping() == nil
	if !driverIsOk {
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
	router.Get("/ping", application.handleCheckDBStatus)

	return router
}
