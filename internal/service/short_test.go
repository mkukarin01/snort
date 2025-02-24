package service

import (
	"testing"

	"github.com/mkukarin01/snort/internal/storage"
	"github.com/stretchr/testify/assert"
)

// проверил что сократилось и получилось тоже самое
func TestURLShortener_ShortenAndRetrieve(t *testing.T) {
	storage := storage.NewMemoryStorage()
	shortener := NewURLShortener(storage)

	originalURL := "https://ya.ru"
	id := shortener.Shorten(originalURL)

	retrievedURL, ok := shortener.Retrieve(id)
	assert.True(t, ok)
	assert.Equal(t, originalURL, retrievedURL)
}

// проверил что получение несуществующего вернет ошибку
func TestURLShortener_RetrieveNonExistent(t *testing.T) {
	storage := storage.NewMemoryStorage()
	shortener := NewURLShortener(storage)

	_, ok := shortener.Retrieve("nonexistent")
	assert.False(t, ok)
}
