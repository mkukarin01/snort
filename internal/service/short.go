package service

import (
	"math/rand"

	"github.com/mkukarin01/snort/internal/storage"
)

// URLShortener обертка для хранилища
type URLShortener struct {
	store storage.Storager
}

// NewURLShortener создаёт новый URLShortener
func NewURLShortener(store storage.Storager) *URLShortener {
	return &URLShortener{store: store}
}

// Shorten создает короткий идентификатор для ссылки
func (us *URLShortener) Shorten(originalURL string) string {
	id := generateID()
	us.store.Save(id, originalURL)
	return id
}

// Shorten создает короткие идентификаторы для ссылок
func (us *URLShortener) ShortenBatch(urls map[string]string) map[string]string {
	result := make(map[string]string)
	batchData := make(map[string]string)

	for correlationID, originalURL := range urls {
		id := generateID()
		result[correlationID] = id
		batchData[id] = originalURL
	}

	// race condition обрабатывается на слое хранилища
	us.store.SaveBatch(batchData)

	return result
}

// Retrieve юзаем стор чтобы вытащить данные по идентификатору и возвращаем + ок
func (us *URLShortener) Retrieve(id string) (string, bool) {
	return us.store.Load(id)
}

// generateID рандомный идентификатор, написал тупую функцию
func generateID() string {
	const length = 8
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// слайс байтов в длину length === 8
	id := make([]byte, length)
	for i := range id {
		// берем случайный индекс
		randomIndex := rand.Intn(len(charset))
		// для позиции i ставим символ из charset
		id[i] = charset[randomIndex]
	}
	return string(id)
}
