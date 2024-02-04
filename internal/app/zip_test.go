package app

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetZipMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusTeapot)
		writer.Write([]byte("blahblahblahblahblahblahblahblahblahblahblahblah"))
	})

	testedHandler := getZipMiddleware(zap.NewNop())(nextHandler)

	noZipRequest := httptest.NewRequest(http.MethodGet, "/some/request", nil)
	noZipWriter := httptest.NewRecorder()
	testedHandler.ServeHTTP(noZipWriter, noZipRequest)

	acceptEncoding := noZipWriter.Header().Get("Content-Encoding")
	assert.Empty(t, acceptEncoding)

	zipRequest := httptest.NewRequest(http.MethodGet, "/some/request", nil)
	zipRequest.Header.Set("Accept-Encoding", "gzip")
	zipWriter := httptest.NewRecorder()
	testedHandler.ServeHTTP(zipWriter, zipRequest)
	acceptEncoding = zipWriter.Header().Get("Content-Encoding")
	assert.Equal(t, "gzip", acceptEncoding)
}
