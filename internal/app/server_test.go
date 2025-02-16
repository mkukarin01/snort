package app

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// тесты разделены по слоям - логика работы с ссылками (БЛ), мидлвари сервака и другая его шелуха
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

// Смотри префикс теста - TestServer - как работает сервак
// проверил пост запрос, чет кривовато наверно
func TestServer_HandlePost(t *testing.T) {
    server := NewServer()

    req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))
    req.Header.Set("Content-Type", "text/plain")
    w := httptest.NewRecorder()

    server.ServeHTTP(w, req)

    res := w.Result()
	// https://habr.com/ru/companies/otus/articles/833702/ отложтл закрытие на чтение, даже в случае ошибки все равно закроемся
    defer res.Body.Close()

    require.Equal(t, http.StatusCreated, res.StatusCode)
    assert.Contains(t, w.Body.String(), "http://localhost:8080/")
}

// проверил гет
func TestServer_HandleGet(t *testing.T) {
    server := NewServer()

    id := server.shortener.Shorten("https://ya.ru")
    url := "/" + id

    req := httptest.NewRequest(http.MethodGet, url, nil)
    w := httptest.NewRecorder()

    server.ServeHTTP(w, req)

    res := w.Result()
    defer res.Body.Close()

    require.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
    assert.Equal(t, "https://ya.ru", res.Header.Get("Location"))
}

// проверил гет на несуществующую ссылку
func TestServer_HandleGet_NonExistent(t *testing.T) {
    server := NewServer()

    req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
    w := httptest.NewRecorder()

    server.ServeHTTP(w, req)

    res := w.Result()
    defer res.Body.Close()

    require.Equal(t, http.StatusBadRequest, res.StatusCode)
}