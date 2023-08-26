package app

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (writer gzipWriter) Write(bytes []byte) (int, error) {
	return writer.Writer.Write(bytes)
}

func getZipMiddleware() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if !strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(writer, request)
				return
			}

			gz, err := gzip.NewWriterLevel(writer, gzip.BestSpeed)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte("Internal error while encode content to gzip: " + err.Error()))
				return
			}

			defer gz.Close()
			writer.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(
				gzipWriter{
					ResponseWriter: writer,
					Writer:         gz,
				},
				request,
			)
		})
	}
}
