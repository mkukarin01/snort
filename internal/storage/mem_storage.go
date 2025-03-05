package storage

import (
	"errors"
	"sync"
)

// memEntry локальная структура хранения
type memEntry struct {
	originalURL string
	userID      string
	isDeleted   bool
}

// MemoryStorage реализация in-memory хранилища
type MemoryStorage struct {
	sync.RWMutex
	store     map[string]*memEntry // short -> данные
	userLinks map[string][]string  // user->[]shortIDs
}

// NewMemoryStorage запускатор "соединения" с памятью, аналогия на NewDatabase
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		store:     make(map[string]*memEntry),
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

func (ms *MemoryStorage) Load(id string) (string, error) {
	ms.RLock()
	defer ms.RUnlock()
	entry, exists := ms.store[id]
	if !exists {
		return "", ErrURLNotFound
	}
	if entry.isDeleted {
		return "", ErrURLDeleted
	}
	return entry.originalURL, nil
}

func (ms *MemoryStorage) FindIDByURL(url string) (string, error) {
	ms.RLock()
	defer ms.RUnlock()
	for shortID, entry := range ms.store {
		if entry.originalURL == url {
			return shortID, nil
		}
	}
	return "", ErrURLNotFound
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
	for existID, entry := range ms.store {
		if entry.originalURL == originalURL && existID != shortID {
			return ErrURLConflict
		}
	}

	// если такой shortID уже есть - это конфликт по short_id
	if oldEntry, ok := ms.store[shortID]; ok && oldEntry.originalURL != originalURL {
		return ErrShortIDConflict
	}

	ms.store[shortID] = &memEntry{
		originalURL: originalURL,
		userID:      userID,
		isDeleted:   false,
	}

	if userID != "" {
		ms.userLinks[userID] = append(ms.userLinks[userID], shortID)
	}

	return nil
}

func (ms *MemoryStorage) SaveBatchUserURLs(userID string, batch map[string]string) error {
	ms.Lock()
	defer ms.Unlock()

	for shortID, originalURL := range batch {
		ms.store[shortID] = &memEntry{
			originalURL: originalURL,
			userID:      userID,
			isDeleted:   false,
		}
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
		entry := ms.store[sid]
		if entry != nil && !entry.isDeleted {
			result = append(result, UserURL{ShortURL: sid, OriginalURL: entry.originalURL})
		}
	}
	return result, nil
}

func (ms *MemoryStorage) MarkUserURLsDeleted(userID string, shortIDs []string) error {
	ms.Lock()
	defer ms.Unlock()

	for _, sid := range shortIDs {
		entry, exists := ms.store[sid]
		if exists && entry.userID == userID {
			entry.isDeleted = true
		}
	}
	return nil
}
