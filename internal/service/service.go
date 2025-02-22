package service

import (
	"math/rand"
	"sync" // читать тут https://pkg.go.dev/sync
)

// URLShortener тип данных сопоставления данных id - ссылка
type URLShortener struct {
	// https://pkg.go.dev/sync#RWMutex
	// хочу чтобы можно было кем угодно читать, но писать одному пока, создал и переиспользуешь на протяжении работы аппы
	sync.RWMutex
	store map[string]string
}

// NewURLShortener создаёт новый экземпляр URLShortener, сделал чтобы меньше писать кода
// вдруг по каким-то причинам захочется разделить потоки данных
func NewURLShortener() *URLShortener {
	return &URLShortener{
		store: make(map[string]string),
	}
}

// Shorten создает короткий идентификатор для ссылки
func (us *URLShortener) Shorten(originalURL string) string {
	id := generateID()
	us.Lock()
	us.store[id] = originalURL
	us.Unlock()
	return id
}

// Retrieve юзаем стор чтобы вытащить данные по идентификатору и возвращаем + ок
func (us *URLShortener) Retrieve(id string) (string, bool) {
	us.RLock()
	url, ok := us.store[id]
	us.RUnlock()
	return url, ok
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
