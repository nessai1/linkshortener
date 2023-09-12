package shortener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nessai1/linkshortener/internal/app"
	"net/http"

	"github.com/go-chi/chi"
)

type AddURLRequestBody struct {
	URL string `json:"url"`
}

type AddURLRequestResult struct {
	Result string `json:"result"`
}

type BatchItemRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BathRequest []BatchItemRequest

type BadRequest struct {
	ErrorMsg string `json:"error_msg"`
}

type BatchItemResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type BatchResponse []BatchItemResponse

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

func (application *Application) apiHandleAddBatchURL(writer http.ResponseWriter, request *http.Request) {
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(request.Body)
	if err != nil {
		application.logger.Debug(fmt.Sprintf("Client sends invalid request. (%s)", err.Error()))
		writer.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	var requestBody BathRequest
	err = json.Unmarshal(buffer.Bytes(), &requestBody)
	if err != nil {
		application.logger.Debug(fmt.Sprintf("Cannot unmarshal client request. (%s)", err.Error()))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := application.SQLDriver.Begin()
	if err != nil {
		s := BadRequest{
			ErrorMsg: fmt.Sprintf("Error while make transaction: %s", err.Error()),
		}

		bt, _ := json.Marshal(s)

		writer.Write(bt)
		writer.WriteHeader(http.StatusInternalServerError)
	}

	responseBody := make(BatchResponse, 0)

	for _, item := range requestBody {
		if !validateURL([]byte(item.OriginalURL)) {
			msg := fmt.Sprintf("Client sends invalid URL \"%s\" in batch item %s.", item.OriginalURL, item.CorrelationID)
			application.logger.Debug(msg)
			errorAnswer := BadRequest{ErrorMsg: msg}
			rs, _ := json.Marshal(errorAnswer)
			writer.Write(rs)
			writer.WriteHeader(http.StatusBadRequest)
			tx.Rollback()
			return
		}

		hash, err := application.createResource(item.OriginalURL)
		if err != nil {
			msg := fmt.Sprintf("Error while create resource \"%s\" in batch item %s. (%s)", item.OriginalURL, item.CorrelationID, err.Error())
			application.logger.Error(msg)
			errorAnswer := BadRequest{ErrorMsg: msg}
			rs, _ := json.Marshal(errorAnswer)
			writer.Write(rs)
			writer.WriteHeader(http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		responseBody = append(responseBody, BatchItemResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      application.buildTokenTail(request) + hash,
		})
	}

	err = tx.Commit()
	if err != nil {
		s := BadRequest{
			ErrorMsg: fmt.Sprintf("Error while commit transaction: %s", err.Error()),
		}

		bt, _ := json.Marshal(s)

		writer.Write(bt)
		writer.WriteHeader(http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")

	requestResult, _ := json.Marshal(responseBody)

	application.logger.Info(fmt.Sprintf("Client success add batch with %d URLs  by API", len(requestBody)))
	writer.WriteHeader(http.StatusCreated)
	writer.Write(requestResult)
}

func (application *Application) getAPIRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(app.GetRequestLogMiddleware(application.logger, "API"))

	router.Post("/", application.apiHandleAddURL)
	router.Post("/batch", application.apiHandleAddBatchURL)

	return router
}
