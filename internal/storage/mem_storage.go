package storage

import (
	"errors"
	"sync"
)

// MemoryStorage реализация in-memory хранилища
type MemoryStorage struct {
	sync.RWMutex
	store     map[string]string   // старое поле: short -> original
	userLinks map[string][]string // user->[]shortIDs
}

// NewMemoryStorage запускатор "соединения" с памятью, аналогия на NewDatabase
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		store:     make(map[string]string),
		userLinks: make(map[string][]string),
	}
}

// -- методы из старого интерфейса --

func (ms *MemoryStorage) Save(id, url string) error {
	return ms.SaveUserURL("", id, url)
}

func (ms *MemoryStorage) SaveBatch(urls map[string]string) error {
	return ms.SaveBatchUserURLs("", urls)
}

func (ms *MemoryStorage) Load(id string) (string, bool) {
	ms.RLock()
	defer ms.RUnlock()
	url, exists := ms.store[id]
	return url, exists
}

func (ms *MemoryStorage) FindIDByURL(url string) (string, bool) {
	ms.RLock()
	defer ms.RUnlock()
	for id, storedURL := range ms.store {
		if storedURL == url {
			return id, true
		}
	}
	return "", false
}

// ну, тут тоже как бы странно было бы закрывать память, но можно че-нить
// по OOM и прочим приколам попробовать реализовать, но наверно, такое не случится
func (ms *MemoryStorage) Ping() error  { return errors.New("there is no connection: mem001") }
func (ms *MemoryStorage) Close() error { return nil }

// -- методы с юид --

func (ms *MemoryStorage) SaveUserURL(userID, shortID, originalURL string) error {
	ms.Lock()
	defer ms.Unlock()

	// если какой-то другой shortID уже хранит этот url - вернём конфликт
	for existID, existURL := range ms.store {
		if existURL == originalURL && existID != shortID {
			return ErrURLConflict
		}
	}

	// если такой shortID уже есть - это конфликт по short_id
	if oldURL, ok := ms.store[shortID]; ok && oldURL != originalURL {
		return ErrShortIDConflict
	}

	ms.store[shortID] = originalURL
	if userID != "" {
		ms.userLinks[userID] = append(ms.userLinks[userID], shortID)
	}

	return nil
}

func (ms *MemoryStorage) SaveBatchUserURLs(userID string, batch map[string]string) error {
	ms.Lock()
	defer ms.Unlock()

	for shortID, originalURL := range batch {
		ms.store[shortID] = originalURL
		if userID != "" {
			ms.userLinks[userID] = append(ms.userLinks[userID], shortID)
		}
	}
	return nil
}

func (ms *MemoryStorage) GetUserURLs(userID string) ([]UserURL, error) {
	ms.RLock()
	defer ms.RUnlock()

	shortIDs, ok := ms.userLinks[userID]
	if !ok {
		return []UserURL{}, nil
	}

	var result []UserURL
	for _, sid := range shortIDs {
		orig := ms.store[sid]
		result = append(result, UserURL{ShortURL: sid, OriginalURL: orig})
	}
	return result, nil
}
