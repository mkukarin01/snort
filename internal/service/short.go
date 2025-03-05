package service

import (
	"errors"
	"math/rand"

	"github.com/mkukarin01/snort/internal/storage"
)

// URLShortener обертка для хранилища
type URLShortener struct {
	store storage.Storager
}

// UserURL структурка (short_url, original_url)
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// NewURLShortener создаёт новый URLShortener
func NewURLShortener(store storage.Storager) *URLShortener {
	return &URLShortener{store: store}
}

// Shorten создает короткий идентификатор для ссылки по userID
// Возвращает сам идентификатор и ошибку (дубликат ссылки или другая проблема)
func (us *URLShortener) Shorten(originalURL, userID string) (string, error) {
	for {
		id := generateID()
		err := us.store.SaveUserURL(userID, id, originalURL)
		if err == nil {
			// успех
			return id, nil
		}

		// ошибка - разрбираемся что происходит
		if errors.Is(err, storage.ErrURLConflict) {
			// все уже в хранилище, найдем другой shortId и вернем 409 или другую ошибку
			existingID, saveErr := us.store.FindIDByURL(originalURL)

			if saveErr == nil {
				return existingID, err
			}

			return "", saveErr
		}

		if errors.Is(err, storage.ErrShortIDConflict) {
			// коллизия по short_id, попробуем сгенерировать заново
			continue
		}

		// Прочие ошибки - завершаем
		return "", err
	}
}

// ShortenBatch создает короткие идентификаторы для ссылок
func (us *URLShortener) ShortenBatch(urls map[string]string, userID string) map[string]string {
	result := make(map[string]string)
	batchData := make(map[string]string)

	for correlationID, originalURL := range urls {
		id := generateID()
		result[correlationID] = id
		batchData[id] = originalURL
	}

	us.store.SaveBatchUserURLs(userID, batchData)

	return result
}

// Retrieve юзаем стор, чтобы вытащить данные по идентификатору и возвращаем + ok
// Если ссылка "удалена", вернем ErrURLDeleted
func (us *URLShortener) Retrieve(id string) (string, error) {
	url, err := us.store.Load(id)
	if err != nil {
		return "", err
	}
	return url, nil
}

// UserURLs возвращает все ссылки по userID
func (us *URLShortener) UserURLs(userID string, baseURL string) ([]UserURL, error) {
	urls, err := us.store.GetUserURLs(userID)
	if err != nil {
		return nil, err
	}
	results := make([]UserURL, 0, len(urls))
	for _, u := range urls {
		results = append(results, UserURL{
			ShortURL:    baseURL + "/" + u.ShortURL,
			OriginalURL: u.OriginalURL,
		})
	}
	return results, nil
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
