package logger

// следуя Clean Architecture каждая внешняя библиотека и ее синглтончики лучше
// реализовавать отдельно от имплементации, таким образом можно написать функцию
// которая будет реализовывать ретрофит апи и новый апи для новой библиотеки
// (банальный ремап аргументов, в другие аргументы, обогащения и тд и тп).
// https://alexkondov.com/full-stack-tao-clean-architecture-react/
// https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// InitLogger инициализирует и сигнлтонит
func InitLogger() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return log
}

// Можно будет вот эту функцию вынести в следующую часть пакета и
// изменить logrus на другую библиотеку

// LoggingMiddleware сама мидлварь, логирует и всякое красивое делает
func LoggingMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			// обернулся и покатился дальше
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			// лог запроса
			logger.WithFields(logrus.Fields{
				"method":        r.Method,
				"uri":           r.RequestURI,
				"status":        ww.statusCode,
				"response_time": duration.Seconds(),
				"response_size": ww.size,
			}).Info("Processed request")
		})
	}
}

// responseWriter структура для записи информации
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
