package shortener

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
			r := httptest.NewRequest(http.MethodPost, testApp.GetAddr(), strings.NewReader(tt.input.body))

			if tt.input.userUUID != "" {
				ctx := context.WithValue(r.Context(), app.ContextUserUUIDKey, app.UserUUID(tt.input.userUUID))
				r = r.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			testApp.apiHandleAddURL(w, r)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
			if tt.expectedLink != nil {
				var buffer bytes.Buffer
				n, err := buffer.ReadFrom(w.Result().Body)
				require.NoError(t, err, "Test expected created link, got error while read body")
				require.NotEqual(t, n, 0, "Test expected created link, got empty result body")

				var result AddURLRequestResult
				err = json.Unmarshal(buffer.Bytes(), &result)
				require.NoError(t, err, "Test expected created link, got error while unmarshal body")
				hash, err := fetchContentAfterTokenTail(result.Result)
				require.NoError(t, err)

				assert.Equal(t, *tt.expectedLink, existingHashes[hash])
			}
		})
	}
}
