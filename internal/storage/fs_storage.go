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
	IsDeleted   bool   `json:"is_deleted"`
}

// FileStorage реализация хранилища в файле
type FileStorage struct {
	sync.RWMutex
	filePath  string
	store     map[string]*fileEntry // short_id -> *fileEntry
	userLinks map[string][]string   // user->[]shortIDs
}

// NewFileStorage запускатор "соединения" с файлом, аналогия на NewDatabase
func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath:  filePath,
		store:     make(map[string]*fileEntry),
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

func (fs *FileStorage) Load(id string) (string, error) {
	fs.RLock()
	defer fs.RUnlock()
	entry, ok := fs.store[id]
	if !ok {
		return "", ErrURLNotFound
	}
	if entry.IsDeleted {
		return "", ErrURLDeleted
	}
	return entry.OriginalURL, nil
}

func (fs *FileStorage) FindIDByURL(url string) (string, bool) {
	fs.RLock()
	defer fs.RUnlock()
	for id, entry := range fs.store {
		if entry.OriginalURL == url {
			return id, true
		}
	}
	return "", false
}

func (fs *FileStorage) Ping() error  { return errors.New("there is no connection: fs001") }
func (fs *FileStorage) Close() error { return nil }

// -- методы с юид --

func (fs *FileStorage) SaveUserURL(userID, shortID, originalURL string) error {
	fs.Lock()
	defer fs.Unlock()

	// если уже есть такой originalURL у другого shortID - вернём конфликт
	for existID, entry := range fs.store {
		if entry.OriginalURL == originalURL && existID != shortID {
			return ErrURLConflict
		}
	}
	// проверка на конфликт shortID
	if oldEntry, ok := fs.store[shortID]; ok && oldEntry.OriginalURL != originalURL {
		return ErrShortIDConflict
	}

	fs.store[shortID] = &fileEntry{
		ShortURL:    shortID,
		OriginalURL: originalURL,
		UserID:      userID,
		IsDeleted:   false,
	}
	if userID != "" {
		fs.userLinks[userID] = append(fs.userLinks[userID], shortID)
	}

	return fs.save()
}

func (fs *FileStorage) SaveBatchUserURLs(userID string, batch map[string]string) error {
	fs.Lock()
	defer fs.Unlock()

	for shortID, originalURL := range batch {
		fs.store[shortID] = &fileEntry{
			ShortURL:    shortID,
			OriginalURL: originalURL,
			UserID:      userID,
			IsDeleted:   false,
		}
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
		entry := fs.store[sid]
		if entry != nil && !entry.IsDeleted {
			result = append(result, UserURL{
				ShortURL:    sid,
				OriginalURL: entry.OriginalURL,
			})
		}
	}
	return result, nil
}

// MarkUserURLsDeleted - множественное обновление для userID и списка shortIDs
func (fs *FileStorage) MarkUserURLsDeleted(userID string, shortIDs []string) error {
	fs.Lock()
	defer fs.Unlock()

	for _, sid := range shortIDs {
		entry, ok := fs.store[sid]
		if ok && entry.UserID == userID {
			entry.IsDeleted = true
		}
	}

	return fs.save()
}

// ----------------- Внутренние методы -----------------

func (fs *FileStorage) save() error {
	file, err := os.Create(fs.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)

	for _, entry := range fs.store {
		if entry == nil {
			continue
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

		fs.store[entry.ShortURL] = &entry
		if entry.UserID != "" {
			fs.userLinks[entry.UserID] = append(fs.userLinks[entry.UserID], entry.ShortURL)
		}
	}
	return nil
}
