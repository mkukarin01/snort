package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync" // читать тут https://pkg.go.dev/sync
)

// FileStorage реализация хранилища в файле
type FileStorage struct {
	// https://pkg.go.dev/sync#RWMutex
	// хочу чтобы можно было кем угодно читать, но писать одному пока, создал и переиспользуешь на протяжении работы аппы
	sync.RWMutex
	filePath string
	store    map[string]string
}

// NewFileStorage запускатор "соединения" с файлом, аналогия на NewDatabase
func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath: filePath,
		store:    make(map[string]string),
	}
	if err := fs.load(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (fs *FileStorage) Save(id, url string) error {
	fs.Lock()
	defer fs.Unlock()
	fs.store[id] = url
	return fs.save()
}

// SaveBatch сохраняем батч
func (fs *FileStorage) SaveBatch(urls map[string]string) error {
	fs.Lock()
	defer fs.Unlock()

	for id, url := range urls {
		fs.store[id] = url
	}

	return fs.save()
}

func (fs *FileStorage) Load(id string) (string, bool) {
	fs.RLock()
	defer fs.RUnlock()
	url, exists := fs.store[id]
	return url, exists
}

func (fs *FileStorage) save() error {
	file, err := os.Create(fs.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	// return enc.Encode(fs.store)
	for id, url := range fs.store {
		entry := map[string]string{"short_url": id, "original_url": url}
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
		entry := make(map[string]string)
		if err := dec.Decode(&entry); err != nil {
			if err.Error() == "EOF" { // Читаем до конца, но не выдаём ошибку если это просто конец файла
				break
			}
			return fmt.Errorf("failed to decode JSON: %w", err)
		}
		fs.store[entry["short_url"]] = entry["original_url"]
	}
	return nil
}

// фс можно оставить без пинга и закрытия
func (fs *FileStorage) Ping() error  { return errors.New("there is no connection: fs001") }
func (fs *FileStorage) Close() error { return nil }
