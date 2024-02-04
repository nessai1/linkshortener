package shortener

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestInitJSONConfig(t *testing.T) {
	_, err := initJSONConfig("someundefinedpath")
	assert.Error(t, err)

	tempFile, err := os.CreateTemp("", "tempconfig")
	configPath := tempFile.Name()

	tempFile.Write([]byte(`{
	  "server_address": "serveraddr",
	  "base_url": "baseurl",
	  "file_storage_path": "pathtostorage",
	  "database_dsn": "",
	  "enable_https": true
	}`))
	require.NoError(t, tempFile.Close())

	config, err := initJSONConfig(configPath)
	require.NoError(t, err)

	assert.Equal(t, "serveraddr", config.ServerAddr)
	assert.Equal(t, "baseurl", config.TokenTail)
	assert.Equal(t, "pathtostorage", config.FileStoragePath)
	assert.Equal(t, "", config.SQLConnection)
	assert.Equal(t, true, config.EnableHTTPS)

	secondTempFile, err := os.CreateTemp("", "emptyconfig")
	require.NoError(t, err)
	configPath = secondTempFile.Name()
	require.NoError(t, secondTempFile.Close())

	_, err = initJSONConfig(configPath)
	assert.Error(t, err)
}

func TestBuildAppConfig(t *testing.T) {
	_, err := BuildAppConfig()
	require.NoError(t, err)
	// разобразться с переопределением флагов и добавить тестов к 9му инкременту
}
