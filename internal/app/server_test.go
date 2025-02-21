package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mkukarin01/snort/internal/config"
)

// тесты разделены по слоям - логика работы с ссылками (БЛ), мидлвари роутера и другая его шелуха
// Смотри префикс теста - TestURLShortener - логика сокращения
// проверил что сократилось и получилось тоже самое
func TestURLShortener_ShortenAndRetrieve(t *testing.T) {
	shortener := NewURLShortener()

	originalURL := "https://ya.ru"
	id := shortener.Shorten(originalURL)

	retrievedURL, ok := shortener.Retrieve(id)
	assert.True(t, ok)
	assert.Equal(t, originalURL, retrievedURL)
}

// проверил что получение несуществующего вернет ошибку
func TestURLShortener_RetrieveNonExistent(t *testing.T) {
	shortener := NewURLShortener()

	_, ok := shortener.Retrieve("nonexistent")
	assert.False(t, ok)
}

// Смотри префикс теста - TestRouter - как работает роутер
// проверил пост запрос, чет кривовато наверно
func TestRouter_HandlePost(t *testing.T) {
	cfg := createCfg()
	router := NewRouter(cfg)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Contains(t, w.Body.String(), "http://localhost:8080/")
}

// проверил гет
func TestRouter_HandleGet(t *testing.T) {
	shortener := NewURLShortener()
	id := shortener.Shorten("https://ya.ru")
	url := "/" + id

	router := createTestRouter(shortener)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, "https://ya.ru", res.Header.Get("Location"))
}

// проверил гет на несуществующую ссылку
func TestRouter_HandleGet_NonExistent(t *testing.T) {
	cfg := createCfg()
	router := NewRouter(cfg)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// создадим тестовый маршрутизатор с URLShortener-ом.
func createTestRouter(shortener *URLShortener) chi.Router {
	r := chi.NewRouter()
	cfg := createCfg()

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handlePost(w, r, shortener, cfg.BaseURL)
	})

	if cfg.BasePath == "" {
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			handleGet(w, r, shortener)
		})
	} else {
		r.Route(cfg.BasePath, func(r chi.Router) {
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				handleGet(w, r, shortener)
			})
		})
	}

	return r
}

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
