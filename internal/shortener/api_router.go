package shortener

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener/encoder"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"go.uber.org/zap"
	"net/http"

	"github.com/go-chi/chi"
)

type AddURLRequestBody struct {
	URL string `json:"url"`
}

type AddURLRequestResult struct {
	Result string `json:"result"`
}

type GetUserURLsResult struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

type BatchItemRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type DeleteUserURLsRequest []string

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
	UserUUID, err := app.RequireUserUUID(request)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		application.logger.Error("Error while authorize user", zap.Error(err))
		return
	}

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(request.Body)
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

	hash, err := application.createResource(linkstorage.Link{
		Value:     requestBody.URL,
		OwnerUUID: &UserUUID,
	})
	if err != nil {

		if errors.Is(err, linkstorage.ErrURLIntersection) {
			writer.WriteHeader(http.StatusConflict)
			application.logger.Debug(fmt.Sprintf("User insert dublicate url: %s", requestBody.URL))
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
			application.logger.Debug(fmt.Sprintf("Cannot create resource for \"%s\". (%s)", requestBody.URL, err.Error()))
			application.logger.Error(fmt.Sprintf("Error while creating resource '%s'\n", requestBody.URL))
			return
		}
	}

	link := application.buildTokenTail(request) + hash

	requestResult, _ := json.Marshal(AddURLRequestResult{Result: link})

	application.logger.Info(fmt.Sprintf("Client success add URL \"%s\" by API", requestBody.URL))
	writer.WriteHeader(http.StatusCreated)
	writer.Write(requestResult)
}

func (application *Application) apiHandleAddBatchURL(writer http.ResponseWriter, request *http.Request) {
	UserUUID, err := app.RequireUserUUID(request)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		application.logger.Error("Error while authorize user", zap.Error(err))
		return
	}

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(request.Body)
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

	innerKWRows := make([]linkstorage.KeyValueRow, len(requestBody))
	expectedResult := make(BatchResponse, len(requestBody))
	for i, item := range requestBody {
		if !validateURL([]byte(item.OriginalURL)) {
			msg := fmt.Sprintf("Client sends invalid URL \"%s\" in batch item %s.", item.OriginalURL, item.CorrelationID)
			application.logger.Debug(msg)
			errorAnswer := BadRequest{ErrorMsg: msg}
			rs, _ := json.Marshal(errorAnswer)
			writer.Write(rs)
			writer.WriteHeader(http.StatusBadRequest)
		}
		hash, err := encoder.EncodeURL(item.OriginalURL)
		if err != nil {
			msg := fmt.Sprintf("Error while hashing URL \"%s\": %s.", item.OriginalURL, err.Error())
			application.logger.Debug(msg)
			errorAnswer := BadRequest{ErrorMsg: msg}
			rs, _ := json.Marshal(errorAnswer)
			writer.Write(rs)
			writer.WriteHeader(http.StatusInternalServerError)
		}

		innerKWRows[i] = linkstorage.KeyValueRow{
			Key:       hash,
			Value:     item.OriginalURL,
			OwnerUUID: &UserUUID,
		}
		expectedResult[i] = BatchItemResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      application.buildTokenTail(request) + hash,
		}
	}

	if err = application.storage.LoadBatch(innerKWRows); err != nil {
		msg := fmt.Sprintf("Error while loading batch: %s.", err.Error())
		application.logger.Debug(msg)
		errorAnswer := BadRequest{ErrorMsg: msg}
		rs, _ := json.Marshal(errorAnswer)
		writer.Write(rs)
		writer.WriteHeader(http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")

	requestResult, _ := json.Marshal(expectedResult)

	application.logger.Info(fmt.Sprintf("Client success add batch with %d URLs  by API", len(requestBody)))
	writer.WriteHeader(http.StatusCreated)
	writer.Write(requestResult)
}

func (application *Application) apiHandleGetUserURLs(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie(app.LoginCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			writer.WriteHeader(http.StatusUnauthorized)
		} else {
			application.logger.Error(fmt.Sprintf("error while get user login cookie: %s", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	UserUUID, err := app.FetchUUID(cookie.Value)
	if err != nil {
		application.logger.Error(fmt.Sprintf("User sends invalid cookie: %s", err.Error()))
		c := &http.Cookie{
			Name:   app.LoginCookieName,
			Value:  "",
			MaxAge: -1,
		}
		http.SetCookie(writer, c)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	result := make([]GetUserURLsResult, 0)
	rows := application.storage.FindByUserUUID(UserUUID)
	if len(rows) == 0 {
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	for _, row := range rows {
		result = append(result, GetUserURLsResult{
			OriginalURL: row.Value,
			ShortURL:    application.buildTokenTail(request) + row.Key,
		})
	}

	rs, _ := json.Marshal(result)
	writer.Write(rs)
}

func (application *Application) apiHandleDeleteURLs(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie(app.LoginCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			writer.WriteHeader(http.StatusUnauthorized)
		} else {
			application.logger.Error(fmt.Sprintf("error while get user login cookie: %s", err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	UserUUID, err := app.FetchUUID(cookie.Value)
	if err != nil {
		application.logger.Error(fmt.Sprintf("User sends invalid cookie: %s", err.Error()))
		c := &http.Cookie{
			Name:   app.LoginCookieName,
			Value:  "",
			MaxAge: -1,
		}
		http.SetCookie(writer, c)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(request.Body)
	if err != nil {
		application.logger.Debug(fmt.Sprintf("Client sends invalid request. (%s)", err.Error()))
		writer.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	var requestBody DeleteUserURLsRequest
	err = json.Unmarshal(buffer.Bytes(), &requestBody)
	if err != nil {
		application.logger.Debug(fmt.Sprintf("Cannot unmarshal client request. (%s)", err.Error()))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	go func() {
		deleteBatch := make([]linkstorage.Hash, 0)
		for _, val := range requestBody {
			deleteBatch = append(deleteBatch, linkstorage.Hash{
				Value:     val,
				OwnerUUID: &UserUUID,
			})
		}

		err := application.storage.DeleteBatch(deleteBatch)
		if err != nil {
			application.logger.Error(fmt.Sprintf("Error while delete links for User %s: %s", UserUUID, err.Error()))
		}
	}()
	writer.WriteHeader(http.StatusAccepted)
}

func (application *Application) getAPIRouter() *chi.Mux {
	linksRouter := chi.NewRouter()
	linksRouter.Use(app.GetRegisterMiddleware(application.logger))
	linksRouter.Post("/", application.apiHandleAddURL)
	linksRouter.Post("/batch", application.apiHandleAddBatchURL)

	userRouter := chi.NewRouter()
	userRouter.Use(app.GetAuthMiddleware(application.logger))
	userRouter.Get("/urls", application.apiHandleGetUserURLs)
	userRouter.Delete("/urls", application.apiHandleDeleteURLs)

	router := chi.NewRouter()
	router.Use(app.GetRequestLogMiddleware(application.logger, "API"))
	router.Mount("/shorten", linksRouter)
	router.Mount("/user", userRouter)

	return router
}
