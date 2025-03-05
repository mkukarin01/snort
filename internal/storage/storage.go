package storage

import (
	"errors"

	"github.com/mkukarin01/snort/internal/config"
)

// пакетные переменные-ошибки для удобства определения типа ошибки
var (
	// ErrDBConnection - проблема связи с бд
	ErrDBConnection = errors.New("db connection issue")
	// ErrURLConflict - original_url уже есть в базе
	ErrURLConflict = errors.New("url conflict")
	// ErrShortIDConflict - короткий short_id уже занят
	ErrShortIDConflict = errors.New("short_id conflict")
	// ErrURLNotFound - нет такой строки в хранилище
	ErrURLNotFound = errors.New("url not found")
	// ErrURLDeleted - is_deleted=true
	ErrURLDeleted = errors.New("url is deleted")
)

// UserURL - для возврата набора ссылок конкретного пользователя
type UserURL struct {
	ShortURL    string
	OriginalURL string
}

// Storager - интерфейс для работы с бд или другим хранилищем
type Storager interface {
	Ping() error
	Close() error

	// Старые методы (без userID)
	Save(id, url string) error
	SaveBatch(urls map[string]string) error
	Load(id string) (string, error)
	FindIDByURL(url string) (string, error)

	// Новые методы для работы с userID
	SaveUserURL(userID, shortID, originalURL string) error
	SaveBatchUserURLs(userID string, batch map[string]string) error
	GetUserURLs(userID string) ([]UserURL, error)

	// Новый метод для проставления флага удаления
	MarkUserURLsDeleted(userID string, shortIDs []string) error
}

// NewStorage определяет используемое хранилище
func NewStorage(cfg *config.Config) (Storager, error) {
	if cfg.DatabaseDSN != "" {
		return NewDatabase(cfg.DatabaseDSN)
	}
	if cfg.FileStoragePath != "" {
		return NewFileStorage(cfg.FileStoragePath)
	}
	return NewMemoryStorage(), nil
}
