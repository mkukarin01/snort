package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// проверил что сократилось и получилось тоже самое
func TestURLShortener_ShortenAndRetrieve(t *testing.T) {
	shortener := NewURLShortener("../../storage.json")

	originalURL := "https://ya.ru"
	id := shortener.Shorten(originalURL)

	retrievedURL, ok := shortener.Retrieve(id)
	assert.True(t, ok)
	assert.Equal(t, originalURL, retrievedURL)
}

// проверил что получение несуществующего вернет ошибку
func TestURLShortener_RetrieveNonExistent(t *testing.T) {
	shortener := NewURLShortener("../../storage.json")

	_, ok := shortener.Retrieve("nonexistent")
	assert.False(t, ok)
}
