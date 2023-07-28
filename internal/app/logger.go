package app

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	length     int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (n int, err error) {
	n, err = lrw.ResponseWriter.Write(b)

	lrw.length += n

	return
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK, 0}
}

func createAppLogger(envType EnvType) (*zap.Logger, error) {
	atom := zap.NewAtomicLevel()

	logLevel, err := getLogLevelByEnvType(envType)
	if err != nil {
		return nil, err
	}

	atom.SetLevel(logLevel)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))

	return logger, nil
}

func getLogLevelByEnvType(envType EnvType) (zapcore.Level, error) {
	if envType == Production {
		return zapcore.ErrorLevel, nil
	} else if envType == Stage {
		return zapcore.InfoLevel, nil
	} else if envType == Development {
		return zapcore.DebugLevel, nil
	}

	return 0, fmt.Errorf("unexpected EnvType got (%d)", envType)
}

func getRequestLogMiddleware(logger *zap.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			startTime := time.Now()

			lrw := NewLoggingResponseWriter(writer)
			next.ServeHTTP(lrw, request)

			duration := time.Since(startTime)

			logger.Info(fmt.Sprintf("Request info: URI = '%s'\tMethod = %s\tDuration = %d", request.RequestURI, request.Method, duration))
			logger.Info(fmt.Sprintf("Response info: URI = '%s'\tStatus = %d\tContent-Length = %d", request.RequestURI, lrw.statusCode, lrw.length))
		})
	}
}
