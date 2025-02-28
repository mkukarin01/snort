package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

// fileEntry структура для сериализации в файл
type fileEntry struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

// FileStorage реализация хранилища в файле
type FileStorage struct {
	sync.RWMutex
	filePath  string
	store     map[string]string   // старое поле: short -> original
	userLinks map[string][]string // user->[]shortIDs
}

// NewFileStorage запускатор "соединения" с файлом, аналогия на NewDatabase
func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath:  filePath,
		store:     make(map[string]string),
		userLinks: make(map[string][]string),
	}
	if err := fs.load(); err != nil {
		return nil, err
	}
	return fs, nil
}

// -- методы из старого интерфейса --

func (fs *FileStorage) Save(id, url string) error {
	return fs.SaveUserURL("", id, url)
}

func (fs *FileStorage) SaveBatch(urls map[string]string) error {
	return fs.SaveBatchUserURLs("", urls)
}

func (fs *FileStorage) Load(id string) (string, bool) {
	fs.RLock()
	defer fs.RUnlock()
	url, exists := fs.store[id]
	return url, exists
}

func (fs *FileStorage) FindIDByURL(url string) (string, bool) {
	fs.RLock()
	defer fs.RUnlock()
	for id, storedURL := range fs.store {
		if storedURL == url {
			return id, true
		}
	}
	return "", false
}

// фс можно оставить без пинга и закрытия
func (fs *FileStorage) Ping() error  { return errors.New("there is no connection: fs001") }
func (fs *FileStorage) Close() error { return nil }

// -- методы с юид --

func (fs *FileStorage) SaveUserURL(userID, shortID, originalURL string) error {
	fs.Lock()
	defer fs.Unlock()

	// если какой-то другой shortID уже хранит этот url - вернём конфликт
	for existID, existURL := range fs.store {
		if existURL == originalURL && existID != shortID {
			return ErrURLConflict
		}
	}
	// проверка на конфликт shortID
	if oldURL, ok := fs.store[shortID]; ok && oldURL != originalURL {
		return ErrShortIDConflict
	}

	fs.store[shortID] = originalURL
	if userID != "" {
		fs.userLinks[userID] = append(fs.userLinks[userID], shortID)
	}

	return fs.save()
}

func (fs *FileStorage) SaveBatchUserURLs(userID string, batch map[string]string) error {
	fs.Lock()
	defer fs.Unlock()

	for shortID, originalURL := range batch {
		fs.store[shortID] = originalURL
		if userID != "" {
			fs.userLinks[userID] = append(fs.userLinks[userID], shortID)
		}
	}

	return fs.save()
}

func (fs *FileStorage) GetUserURLs(userID string) ([]UserURL, error) {
	fs.RLock()
	defer fs.RUnlock()

	shortIDs, ok := fs.userLinks[userID]
	if !ok {
		return []UserURL{}, nil // нет ссылок у этого юид
	}

	var result []UserURL
	for _, sid := range shortIDs {
		orig := fs.store[sid]
		result = append(result, UserURL{ShortURL: sid, OriginalURL: orig})
	}
	return result, nil
}

// ----------------- Внутренние методы -----------------

func (fs *FileStorage) save() error {
	file, err := os.Create(fs.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)

	// shortID -> userID
	shortToUser := make(map[string]string)
	for uid, sids := range fs.userLinks {
		for _, sid := range sids {
			if _, found := shortToUser[sid]; !found {
				shortToUser[sid] = uid
			}
		}
	}

	for shortID, originalURL := range fs.store {
		entry := fileEntry{
			ShortURL:    shortID,
			OriginalURL: originalURL,
			UserID:      shortToUser[shortID],
		}
		if err := enc.Encode(entry); err != nil {
			fmt.Printf("Failed to encode entry: %v\n", err)
		}
	}

	return nil
}

func (fs *FileStorage) load() error {
	file, err := os.Open(fs.filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	for {
		var entry fileEntry
		if err := dec.Decode(&entry); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to decode JSON: %w", err)
		}

		fs.store[entry.ShortURL] = entry.OriginalURL
		if entry.UserID != "" {
			fs.userLinks[entry.UserID] = append(fs.userLinks[entry.UserID], entry.ShortURL)
		}
	}
	return nil
}
