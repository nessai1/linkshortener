package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAppLogger(t *testing.T) {
	tests := []struct {
		name          string
		envType       EnvType
		expectedLevel zapcore.Level
		hasErr        bool
	}{
		{
			name:          "Production",
			envType:       Production,
			expectedLevel: zapcore.ErrorLevel,
		},
		{
			name:          "Stage",
			envType:       Stage,
			expectedLevel: zapcore.InfoLevel,
		},
		{
			name:          "Dev",
			envType:       Development,
			expectedLevel: zapcore.DebugLevel,
		},
		{
			name:    "Undefined",
			envType: EnvType(42),
			hasErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := CreateAppLogger(tt.envType)
			if tt.hasErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedLevel, logger.Level())
		})
	}
}

func TestLoggingResponseWriter(t *testing.T) {
	rw := NewLoggingResponseWriter(httptest.NewRecorder())

	assert.Equal(t, false, rw.headerIsWrite)
	n, err := rw.Write([]byte("abc"))
	require.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, 3, rw.length)

	rw.WriteHeader(http.StatusConflict)
	assert.Equal(t, true, rw.headerIsWrite)
	assert.Equal(t, http.StatusConflict, rw.statusCode)
	rw.WriteHeader(http.StatusForbidden)
	assert.Equal(t, http.StatusConflict, rw.statusCode)
}
