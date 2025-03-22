package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mkukarin01/snort/internal/config"
	"github.com/mkukarin01/snort/internal/service"
	"github.com/mkukarin01/snort/internal/storage"
)

// **** utility ****//

// createTestRouter создаем тестовый роутер
func createTestRouter(s *service.URLShortener) http.Handler {
	cfg := &config.Config{
		Port:       "8080",
		BaseDomain: "localhost",
		BasePath:   "",
		Address:    "localhost:8080",
		BaseURL:    "http://localhost:8080",
	}

	var r http.Handler
	if s != nil {
		r = createRouter(s, cfg)
	} else {
		ms := storage.NewMemoryStorage()
		shortener := service.NewURLShortener(ms)
		r = createRouter(shortener, cfg)
	}

	return r
}

// createShortenedRouter должна повторять структуру оригинального роутера, пока ручками
func createRouter(s *service.URLShortener, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		HandleShorten(w, r, s, cfg.BaseURL)
	})

	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		HandleShortenJSON(w, r, s, cfg.BaseURL)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		HandleRedirect(w, r, s)
	})

	return r
}

// **** cases ****//

// тест ShortenerJSON POST /api/shorten
func TestHandler_ShortenerJSON(t *testing.T) {
	r := createTestRouter(nil)

	requestBody, _ := json.Marshal(URLRequest{
		URL: "https://ya.ru",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)

	var resp URLResponse
	err := json.NewDecoder(res.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Result, "http://localhost:8080/")
}

// Проверка ShortenerPlain POST /
func TestHandler_ShortenerPlain(t *testing.T) {
	r := createTestRouter(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)
}

// Проверка GetShortURL GET /{id}
func TestHandler_GetShortURL(t *testing.T) {
	ms := storage.NewMemoryStorage()
	shortener := service.NewURLShortener(ms)
	uid := "foo"
	id, _ := shortener.Shorten("https://ya.ru", uid)
	url := "/" + id

	r := createTestRouter(shortener)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, "https://ya.ru", res.Header.Get("Location"))
}

// Проверка TestHandlePing_Success GET /ping 200
func TestHandlePing_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// мокаем бд
	mockDB := storage.NewMockStorager(ctrl)
	mockDB.EXPECT().Ping().Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	HandlePing(rec, req, mockDB)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// Проверка TestHandlePing_NoDB GET /ping 500
func TestHandlePing_NoDB(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	HandlePing(rec, req, nil)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

// Проверка TestHandlePing_FailedPing GET /ping 500
func TestHandlePing_FailedPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := storage.NewMockStorager(ctrl)
	mockDB.EXPECT().Ping().Return(assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	HandlePing(rec, req, mockDB)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
