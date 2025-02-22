package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestLogger_Middleware проверяет работу мидлвари
func TestLogger_Middleware(t *testing.T) {
	logBuffer := new(bytes.Buffer)
	log := logrus.New()
	log.SetOutput(logBuffer)

	// тестовый обработчик, просто отдаёт 200 OK.
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// типа каррирование для обертки
	loggedHandler := LoggingMiddleware(log)(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	// буфер как строка и ассертим если там нужные куски данных
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Processed request")
	assert.Contains(t, logOutput, "method=GET")
	assert.Contains(t, logOutput, "status=200")
}
