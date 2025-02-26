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

// Storager - интерфейс для работы с бд или другим хранилищем
type Storager interface {
	Ping() error
	Close() error
	Save(id, url string) error
	SaveBatch(urls map[string]string) error
	Load(id string) (string, bool)
	FindIDByURL(url string) (string, bool)
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
