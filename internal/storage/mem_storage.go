package storage

import (
	"errors"
	"sync"
)

// MemoryStorage реализация in-memory хранилища
type MemoryStorage struct {
	sync.RWMutex
	store map[string]string
}

// NewMemoryStorage запускатор "соединения" с памятью, аналогия на NewDatabase
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		store: make(map[string]string),
	}
}

func (ms *MemoryStorage) Save(id, url string) error {
	ms.Lock()
	defer ms.Unlock()

	// если какой-то другой shortID уже хранит этот url - вернём конфликт
	for existID, existURL := range ms.store {
		if existURL == url {
			if existID != id {
				return ErrURLConflict
			}
		}
	}

	// если такой shortID уже есть - это конфликт по short_id
	if _, ok := ms.store[id]; ok {
		if ms.store[id] != url {
			return ErrShortIDConflict
		}
	}

	ms.store[id] = url
	return nil
}

func (ms *MemoryStorage) SaveBatch(urls map[string]string) error {
	ms.Lock()
	defer ms.Unlock()

	for id, url := range urls {
		ms.store[id] = url
	}

	return nil
}

func (ms *MemoryStorage) Load(id string) (string, bool) {
	ms.RLock()
	defer ms.RUnlock()
	url, exists := ms.store[id]
	return url, exists
}

// FindIDByURL находим short_id по original_url
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
