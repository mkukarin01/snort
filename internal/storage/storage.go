package storage

import (
	"errors"

	"github.com/mkukarin01/snort/internal/config"
)

// пакетные переменные-ошибки для удобства определения типа ошибки
var (
	// ErrDbConnection - проблема связи с бд
	ErrDbConnection = errors.New("db connection issue")
	// ErrURLConflict - original_url уже есть в базе
	ErrURLConflict = errors.New("url conflict")
	// ErrShortIDConflict - короткий short_id уже занят
	ErrShortIDConflict = errors.New("short_id conflict")
	// ErrURLNotFound - нет такой строки в хранилище
	ErrURLNotFound = errors.New("url not found")
)

// Storager - интерфейс для работы с бд или другим хранилищем
type Storager interface {
	Ping() error
	Close() error
	Save(id, url string) error
	SaveBatch(urls map[string]string) error
	Load(id string) (string, error)
	FindIDByURL(url string) (string, error)
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
