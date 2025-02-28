package storage

import (
	"errors"

	"github.com/mkukarin01/snort/internal/config"
)

// пакетные переменные-ошибки для удобства определения типа ошибки
var (
	// ErrURLConflict - original_url уже есть в базе
	ErrURLConflict = errors.New("url conflict")
	// ErrShortIDConflict - короткий short_id уже занят
	ErrShortIDConflict = errors.New("short_id conflict")
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
	Load(id string) (string, bool)
	FindIDByURL(url string) (string, bool)

	// Новые методы для работы с userID
	SaveUserURL(userID, shortID, originalURL string) error
	SaveBatchUserURLs(userID string, batch map[string]string) error
	GetUserURLs(userID string) ([]UserURL, error)
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
