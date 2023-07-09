package shortener

import (
	encoder "github.com/nessai1/linkshortener/internal/shortener/decoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApplication_handleAddURL(t *testing.T) {
	type request struct {
		status int
		body   string
	}

	app := GetApplication()
	serviceURL := "http://" + app.GetAddr() + "/"

	testHash, err := encoder.EncodeURL("https://ya.ru")
	require.NoError(t, err, "Error while encoding test url")

	tests := []struct {
		name          string
		wantedRequest request
		addr          string
	}{
		{
			name: "New addr",
			addr: "https://ya.ru",
			wantedRequest: request{
				status: http.StatusCreated,
				body:   serviceURL + testHash,
			},
		},
		// Проверяем, что при повторной попытке записать адрес - отдает тот же ответ
		{
			name: "Existing addr",
			addr: "https://ya.ru",
			wantedRequest: request{
				status: http.StatusCreated,
				body:   serviceURL + testHash,
			},
		},
		{
			name: "No pattern addr",
			addr: "ftp:mail.ru",
			wantedRequest: request{
				status: http.StatusBadRequest,
				body:   "Invalid pattern of given URI",
			},
		},
		{
			name: "empty addr",
			addr: "",
			wantedRequest: request{
				status: http.StatusBadRequest,
				body:   "Invalid pattern of given URI",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, serviceURL, strings.NewReader(tt.addr))
			w := httptest.NewRecorder()
			app.handleAddURL(w, r)
			res := w.Result()

			assert.Equalf(t, tt.wantedRequest.status, res.StatusCode,
				"Invalid response status %s (%s expected)", res.StatusCode, tt.wantedRequest.status,
			)

			resBody, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			require.NoError(t, err)
			assert.Equalf(t, tt.wantedRequest.body, string(resBody),
				"Invalid response body %s (%s expected)", resBody, tt.wantedRequest.body,
			)
		})
	}
}

func TestApplication_handleGetURL(t *testing.T) {
	type request struct {
		status   int
		location string
	}

	app := GetApplication()
	serviceURL := "http://" + app.GetAddr() + "/"
	testURL := "https://ya.ru"
	testHash, err := app.createResource(testURL)
	require.NoError(t, err, "Error while create test url")

	tests := []struct {
		name          string
		wantedRequest request
		hash          string
	}{
		{
			name: "Get existing resource",
			wantedRequest: request{
				status:   http.StatusTemporaryRedirect,
				location: testURL,
			},
			hash: testHash,
		},
		{
			name: "Get non-existing resource",
			wantedRequest: request{
				status: http.StatusNotFound,
			},
			hash: "NoNeXiSt42",
		},
		{
			name: "Get empty path",
			wantedRequest: request{
				status: http.StatusNotFound,
			},
			hash: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, serviceURL+tt.hash, nil)
			w := httptest.NewRecorder()
			app.handleGetURL(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equalf(t, tt.wantedRequest.status, res.StatusCode,
				"Invalid response status %s (%s expected)", res.StatusCode, tt.wantedRequest.status,
			)

			assert.Equalf(t, tt.wantedRequest.location, res.Header.Get("Location"),
				"Invalid location header %s (%s expected)", res.Header.Get("Location"), tt.wantedRequest.location,
			)
		})
	}
}
