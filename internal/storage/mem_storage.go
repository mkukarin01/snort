package storage

import (
	"errors"
	"sync" // читать тут https://pkg.go.dev/sync
)

// MemoryStorage реализация in-memory хранилища
type MemoryStorage struct {
	// https://pkg.go.dev/sync#RWMutex
	// хочу чтобы можно было кем угодно читать, но писать одному пока, создал и переиспользуешь на протяжении работы аппы
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
	ms.store[id] = url
	return nil
}

func (ms *MemoryStorage) Load(id string) (string, bool) {
	ms.RLock()
	defer ms.RUnlock()
	url, exists := ms.store[id]
	return url, exists
}

// ну, тут тоже как бы странно было бы закрывать память, но можно че-нить
// по OOM и прочим приколам попробовать реализовать, но наверно, такое не случится
func (ms *MemoryStorage) Ping() error  { return errors.New("there is no connection: mem001") }
func (ms *MemoryStorage) Close() error { return nil }
