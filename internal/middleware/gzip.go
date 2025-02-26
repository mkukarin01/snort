package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipResponseWriter обертка для ResponseWriter, позволяющая прозрачно сжимать вывод
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write реализует интерфейс ResponseWriter для gzipResponseWriter
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipMiddleware - мидлварь для обработки сжатия gzip
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		next.ServeHTTP(gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

// GzipDecompressionMiddleware - мидлварь для обработки входящего сжатого трафика
func GzipDecompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			// если пустой, надо проверить и ничего не делать, для этого есть тест
			if r.Body == nil || r.ContentLength == 0 {
				next.ServeHTTP(w, r) // Просто передаем дальше
				return
			}

			// тело есть -> разжимаем
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz // подмена тела на поток
		}

		next.ServeHTTP(w, r)
	})
}
