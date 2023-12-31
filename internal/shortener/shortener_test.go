package shortener

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
)

const TestTokenTail = "test"

func fetchContentAfterTokenTail(link string) (string, error) {
	after, ok := strings.CutPrefix(link, TestTokenTail+"/")
	if !ok {
		return "", fmt.Errorf("cannot find token tail in given string")
	}

	return after, nil
}

func createTestApp(serverAddr, tokenTail string, hl linkstorage.HashToLink) *Application {
	storage := linkstorage.NewMemoryStorage(hl)
	cfg := Config{
		ServerAddr:  serverAddr,
		TokenTail:   tokenTail,
		LinkStorage: storage,
	}

	return &Application{
		config:  &cfg,
		logger:  zap.NewNop(),
		storage: storage,
	}
}

func addChiURLParams(r *http.Request, params map[string]string) *http.Request {
	ctx := chi.NewRouteContext()
	for k, v := range params {
		ctx.URLParams.Add(k, v)
	}

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}

func TestApplication_GetAddr(t *testing.T) {
	tests := []struct {
		name      string
		innerAddr string
		expected  string
	}{
		{
			name:      "Some address",
			innerAddr: "bestlink.com",
			expected:  "bestlink.com",
		},
		{
			name:      "Empty address",
			innerAddr: "",
			expected:  "localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := createTestApp(tt.innerAddr, "", map[string]linkstorage.Link{})
			assert.Equal(t, tt.expected, app.GetAddr())
		})
	}
}

func TestApplication_validateURL(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		isValid bool
	}{
		{
			name:    "Secure address",
			addr:    "https://hello-world.com",
			isValid: true,
		},
		{
			name:    "Insecure address",
			addr:    "http://hello-world.com",
			isValid: true,
		},
		{
			name:    "Invalid address #1",
			addr:    "httpsa://hello-world.com",
			isValid: false,
		},
		{
			name:    "Invalid address #2",
			addr:    "ftp://hello-world.com",
			isValid: false,
		},
		{
			name:    "Invalid address #3",
			addr:    "hello-world.com",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, validateURL([]byte(tt.addr)))
		})
	}
}
