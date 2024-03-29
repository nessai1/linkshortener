package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingResponseWriter декоратор ResponseWriter для логирования доп инфоррмации о запросе
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	length        int
	headerIsWrite bool
}

// WriteHeader обертка базового метода для записи информации о заголовках в лог
func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	if lrw.headerIsWrite {
		return
	}
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
	lrw.headerIsWrite = true
}

// Write обертка базового метода для записи информации о длинне тела запроса в лог
func (lrw *LoggingResponseWriter) Write(b []byte) (n int, err error) {
	n, err = lrw.ResponseWriter.Write(b)

	lrw.length += n

	return
}

// NewLoggingResponseWriter создает обертку LoggingResponseWriter над ResponseWriter
func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, http.StatusOK, 0, false}
}

// CreateAppLogger создает логгер zap.Logger исходя из режима работы приложения EnvType
func CreateAppLogger(envType EnvType) (*zap.Logger, error) {
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

// GetRequestLogMiddleware является middleware который логирует информацию о запросах, которые проходят через методы покрытые этим middleware
func GetRequestLogMiddleware(logger *zap.Logger, prefix string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			startTime := time.Now()

			lrw := NewLoggingResponseWriter(writer)
			next.ServeHTTP(lrw, request)

			duration := time.Since(startTime)

			signCookie, err := request.Cookie(LoginCookieName)
			var userUUID UserUUID
			if err != nil {
				userUUID = "undefined"
			} else {
				userUUID, err = FetchUUID(signCookie.Value)
				if err != nil {
					userUUID = "undefined"
				}
			}

			logger.Info(
				fmt.Sprintf("[%s] Request info", prefix),
				zap.String("URI", request.RequestURI),
				zap.String("Method", request.Method),
				zap.Int64("Duration", int64(duration)),
				zap.String("User UUID", string(userUUID)),
			)

			logger.Info(
				fmt.Sprintf("[%s] Request info", prefix),
				zap.String("URI", request.RequestURI),
				zap.Int("Status", lrw.statusCode),
				zap.Int("Content-Length", lrw.length),
			)
		})
	}
}
