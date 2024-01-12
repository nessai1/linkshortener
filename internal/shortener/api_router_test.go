package shortener

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener/encoder"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type InputDataJSON struct {
	body     string
	userUUID string
}

func TestApplication_apiHandleAddURL(t *testing.T) {
	ownerUUID := "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

	existingHashes := make(linkstorage.HashToLink, 2)
	testApp := createTestApp("localhost:8080", TestTokenTail, existingHashes)

	tests := []struct {
		name           string
		input          InputDataJSON
		expectedStatus int
		expectedLink   *linkstorage.Link
	}{
		{
			name: "Successful create",
			input: InputDataJSON{
				body:     `{"url": "https://yandex.ru"}`,
				userUUID: ownerUUID,
			},
			expectedStatus: http.StatusCreated,
			expectedLink: &linkstorage.Link{
				Value:     "https://yandex.ru",
				OwnerUUID: ownerUUID,
				IsDeleted: false,
			},
		},
		{
			name: "Unsuccessful create: already exists link",
			input: InputDataJSON{
				body:     `{"url": "https://yandex.ru"}`,
				userUUID: ownerUUID,
			},
			expectedStatus: http.StatusConflict,
			expectedLink: &linkstorage.Link{
				Value:     "https://yandex.ru",
				OwnerUUID: ownerUUID,
				IsDeleted: false,
			},
		},
		{
			name: "Unsuccessful create: already exists link",
			input: InputDataJSON{
				body:     `{"url": "https://yandex.ru"}`,
				userUUID: ownerUUID,
			},
			expectedStatus: http.StatusConflict,
			expectedLink: &linkstorage.Link{
				Value:     "https://yandex.ru",
				OwnerUUID: ownerUUID,
				IsDeleted: false,
			},
		},
		{
			name: "Unsuccessful create: bad url",
			input: InputDataJSON{
				body:     `{"url": "ftp://some-filesystem"}`,
				userUUID: ownerUUID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedLink:   nil,
		},
		{
			name: "Unsuccessful create: bad json",
			input: InputDataJSON{
				body:     `some unmarshalling shit`,
				userUUID: ownerUUID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedLink:   nil,
		},
		{
			name: "Unsuccessful create: not authorize",
			input: InputDataJSON{
				body:     `{"url": "https://yandex.ru"}`,
				userUUID: "",
			},
			expectedStatus: http.StatusForbidden,
			expectedLink:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, testApp.GetAddr()+"/api/shorten", strings.NewReader(tt.input.body))

			if tt.input.userUUID != "" {
				ctx := context.WithValue(r.Context(), app.ContextUserUUIDKey, app.UserUUID(tt.input.userUUID))
				r = r.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			testApp.apiHandleAddURL(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			if tt.expectedLink != nil {
				var buffer bytes.Buffer
				n, err := buffer.ReadFrom(res.Body)
				require.NoError(t, err, "Test expected created link, got error while read body")
				require.NotEqual(t, n, 0, "Test expected created link, got empty result body")

				var result addURLRequestResult
				err = json.Unmarshal(buffer.Bytes(), &result)
				require.NoError(t, err, "Test expected created link, got error while unmarshal body")
				hash, err := fetchContentAfterTokenTail(result.Result)
				require.NoError(t, err)

				assert.Equal(t, *tt.expectedLink, existingHashes[hash])
			}
		})
	}
}

func TestApplication_apiHandleAddBatchURL(t *testing.T) {
	ownerUUID := "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

	existingHashes := make(linkstorage.HashToLink, 2)
	testApp := createTestApp("localhost:8080", TestTokenTail, existingHashes)

	vkHash, _ := encoder.EncodeURL("https://vk.com")
	yaHash, _ := encoder.EncodeURL("https://ya.ru")

	tests := []struct {
		name           string
		input          InputDataJSON
		expectedStatus int
		expectedHashes map[string]string // correlation id -> link
	}{
		{
			name: "Successful create: one link",
			input: InputDataJSON{
				body:     `[{"correlation_id": "a", "original_url": "https://vk.com"}]`,
				userUUID: ownerUUID,
			},
			expectedStatus: http.StatusCreated,
			expectedHashes: map[string]string{
				"a": vkHash,
			},
		},
		{
			name: "Successful create: two links",
			input: InputDataJSON{
				body:     `[{"correlation_id": "a", "original_url": "https://vk.com"},{"correlation_id": "b", "original_url": "https://ya.ru"}]`,
				userUUID: ownerUUID,
			},
			expectedStatus: http.StatusCreated,
			expectedHashes: map[string]string{
				"a": vkHash,
				"b": yaHash,
			},
		},
		{
			name: "Unsuccessful create: not authorized",
			input: InputDataJSON{
				body:     `[{"correlation_id": "a", "original_url": "https://vk.com"}]`,
				userUUID: "",
			},
			expectedStatus: http.StatusForbidden,
			expectedHashes: nil,
		},
		{
			name: "Unsuccessful create: incorrect format",
			input: InputDataJSON{
				body:     `[{"correlation_id": "a", "original_url": "https://vk.com"]`,
				userUUID: ownerUUID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedHashes: nil,
		},
		{
			name: "Unsuccessful create: empty body",
			input: InputDataJSON{
				body:     ``,
				userUUID: ownerUUID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedHashes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(tt.input.body))
			if tt.input.userUUID != "" {
				ctx := context.WithValue(request.Context(), app.ContextUserUUIDKey, app.UserUUID(tt.input.userUUID))
				request = request.WithContext(ctx)
			}

			writer := httptest.NewRecorder()
			testApp.apiHandleAddBatchURL(writer, request)
			res := writer.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			if tt.expectedHashes != nil {
				var buffer bytes.Buffer
				n, err := buffer.ReadFrom(writer.Body)
				require.NoError(t, err, "Test expected created list, got error while read")
				require.NotEqual(t, 0, n, "Test expected created list, got empty body")
				innerBody := buffer.String()

				var result batchResponse
				err = json.Unmarshal(buffer.Bytes(), &result)
				require.NoErrorf(t, err, "Test expected created list, got error while unmarshal result body, body: %s", innerBody)

				assert.Equal(t, len(tt.expectedHashes), len(result), "Len of expected hashes and actual must be equal")
				for _, item := range result {
					hash, err := fetchContentAfterTokenTail(item.ShortURL)
					if err != nil {
						assert.NoError(t, err, "Error while get hash from result link")
					} else {
						assert.Equal(t, tt.expectedHashes[item.CorrelationID], hash)
					}
				}
			}
		})
	}
}

func TestApplication_apiHandleDeleteURLs(t *testing.T) {
	ownerUUID := "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

	testApp := createTestApp(":8080", "http://test", nil)

	tests := []struct {
		name           string
		input          InputDataJSON
		expectedStatus int
	}{
		{
			name:           "Successful delete #1",
			input:          InputDataJSON{body: `["abvg"]`, userUUID: ownerUUID},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Invalid body #1",
			input:          InputDataJSON{body: `["abvg" "de"]`, userUUID: ownerUUID},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid body #2",
			input:          InputDataJSON{body: `hello world`, userUUID: ownerUUID},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized",
			input:          InputDataJSON{body: `["hello", "world"]`, userUUID: ""},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, "/user/urls", strings.NewReader(tt.input.body))

			if tt.input.userUUID != "" {
				ctx := request.Context()
				request = request.WithContext(context.WithValue(ctx, app.ContextUserUUIDKey, app.UserUUID(tt.input.userUUID)))
			}

			writer := httptest.NewRecorder()

			testApp.apiHandleDeleteURLs(writer, request)

			assert.Equal(t, tt.expectedStatus, writer.Result().StatusCode)
		})
	}
}
