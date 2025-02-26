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
	id, _ := shortener.Shorten(originalURL)

	retrievedURL, ok := shortener.Retrieve(id)
	assert.True(t, ok)
	assert.Equal(t, originalURL, retrievedURL)

	nextId, conflict := shortener.Shorten(originalURL)
	assert.True(t, conflict)
	assert.Equal(t, id, nextId)
}

// проверил что получение несуществующего вернет ошибку
func TestURLShortener_RetrieveNonExistent(t *testing.T) {
	storage := storage.NewMemoryStorage()
	shortener := NewURLShortener(storage)

	_, ok := shortener.Retrieve("nonexistent")
	assert.False(t, ok)
}

func TestURLShortener_ShortenBatchAndRetrieve(t *testing.T) {
	storage := storage.NewMemoryStorage()
	shortener := NewURLShortener(storage)

	urls := map[string]string{
		"1": "http://ya.ru",
		"2": "http://github.com",
	}

	shortened := shortener.ShortenBatch(urls)

	assert.Len(t, shortened, len(urls))

	for correlationID, originalURL := range urls {
		shortID, exists := shortened[correlationID]
		assert.True(t, exists, "The correlation ID should exist in shortened URLs")

		retrievedURL, ok := shortener.Retrieve(shortID)
		assert.True(t, ok, "Shortened ID should be retrievable")
		assert.Equal(t, originalURL, retrievedURL, "Retrieved URL should match original URL")
	}
}
