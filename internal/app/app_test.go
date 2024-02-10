package app

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestBuildSecureServer(t *testing.T) {
	serverAddr := "serveraddr.com"
	mux := http.DefaultServeMux
	server := buildSecureServer(serverAddr, mux)

	assert.Equal(t, mux, server.Handler)
	assert.Equal(t, serverAddr, server.Addr)
}
