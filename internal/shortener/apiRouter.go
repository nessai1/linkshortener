package shortener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
)

type AddURLRequestBody struct {
	URL string `json:"url"`
}

type AddURLRequestResult struct {
	Result string `json:"result"`
}

type BadRequest struct {
	ErrorMsg string `json:"error_msg"`
}

func (application *Application) apiHandleAddURL(writer http.ResponseWriter, request *http.Request) {
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(request.Body)
	if err != nil {
		application.logger.Debug(fmt.Sprintf("Client sends invalid request. (%s)", err.Error()))
		writer.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	var requestBody AddURLRequestBody
	err = json.Unmarshal(buffer.Bytes(), &requestBody)
	if err != nil {
		application.logger.Debug(fmt.Sprintf("Cannot unmarshal client request. (%s)", err.Error()))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	writer.Header().Set("Content-Type", "application/json")

	if !validateURL([]byte(requestBody.URL)) {
		application.logger.Debug(fmt.Sprintf("Client sends invalid URL \"%s\".", requestBody.URL))
		errorAnswer := BadRequest{ErrorMsg: "Invalid pattern of URL"}
		rs, _ := json.Marshal(errorAnswer)
		writer.Write(rs)
		writer.WriteHeader(http.StatusBadRequest)
	}

	hash, err := application.createResource(requestBody.URL)
	if err != nil {
		application.logger.Debug(fmt.Sprintf("Cannot create resource for \"%s\". %s", requestBody.URL, err.Error()))
		errorAnswer := BadRequest{ErrorMsg: fmt.Sprintf("Error while creating resource '%s'", requestBody.URL)}
		rs, _ := json.Marshal(errorAnswer)
		writer.Write(rs)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	link := application.buildTokenTail(request) + hash

	requestResult, _ := json.Marshal(AddURLRequestResult{Result: link})

	application.logger.Info(fmt.Sprintf("Client success add URL \"%s\" by API", requestBody.URL))
	writer.WriteHeader(http.StatusCreated)
	writer.Write(requestResult)
}

func (application *Application) getAPIRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/some", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("this is api controller"))
	})

	return router
}
