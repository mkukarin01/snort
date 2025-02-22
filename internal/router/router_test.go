package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mkukarin01/snort/internal/config"
)

// **** utility ****//

// createCfg просто конфиг вернем
func createCfg() *config.Config {
	cfg := &config.Config{
		Port:       "8080",
		BaseDomain: "localhost",
		BasePath:   "",
		Address:    "localhost:8080",
		BaseURL:    "http://localhost:8080",
	}
	return cfg
}

// **** cases ****//

// TestRouter_Routes проверяет, что роутер создаётся и все маршруты присутствуют.
func TestRouter_Routes(t *testing.T) {
	cfg := createCfg()
	r := NewRouter(cfg)

	testCases := []struct {
		method string
		path   string
	}{
		{"POST", "/"},
		{"POST", "/api/shorten"},
		{"GET", "/anyShortID"},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		res := w.Result()
		require.NotEqual(t, http.StatusNotFound, res.StatusCode, "Expected route to be registered: %s %s", tc.method, tc.path)
	}
}
