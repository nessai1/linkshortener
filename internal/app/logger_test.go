package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
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

func TestGetRequestLogMiddleware(t *testing.T) {

	tests := []struct {
		name string

		uri      string
		method   string
		userUUID string
	}{
		{
			name: "Request with UUID",

			uri:      "http://test.com/some/path",
			method:   http.MethodGet,
			userUUID: GenerateUserUUID(),
		},
		{
			name:     "Request without UUID",
			uri:      "http://test.com/some/cool/path",
			method:   http.MethodPost,
			userUUID: "",
		},
	}

	testPrefix := "TEST_PREFIX"
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathname := t.TempDir()
			tempFile, err := os.Create(pathname + "/case_" + strconv.Itoa(i))
			tempFileName := tempFile.Name()
			require.NoError(t, err)

			defaultStdout := os.Stdout
			os.Stdout = tempFile

			nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusTeapot)
			})

			logger, _ := CreateAppLogger(Development)
			handlerToTest := GetRequestLogMiddleware(logger, testPrefix)(nextHandler)

			req := httptest.NewRequest(tt.method, tt.uri, nil)
			if tt.userUUID != "" {
				sign, _ := generateSign(tt.userUUID)
				req.AddCookie(&http.Cookie{Name: LoginCookieName, Value: sign})
			}

			handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
			err = tempFile.Close()
			os.Stdout = defaultStdout
			require.NoError(t, err)
			tempFile, err = os.Open(tempFileName)

			out := make([]byte, 1024)
			tempFile.Read(out)

			userUUID := tt.userUUID
			if userUUID == "" {
				userUUID = "undefined"
			}

			innerRequestLogPattern := `info\s\[` + testPrefix + `\]\sRequest info\s{"URI": "` + tt.uri + `", "Method": "` + tt.method + `", "Duration": \d*, "User UUID": "` + userUUID + `"}`

			matched, _ := regexp.MatchString(innerRequestLogPattern, string(out))
			assert.True(t, matched)

			require.NoError(t, err)
		})
	}
}
