package storage

import (
	"github.com/mkukarin01/snort/internal/config"
)

// Storager - интерфейс для работы с бд или другим хранилищем
type Storager interface {
	Ping() error
	Close() error
	Save(id, url string) error
	SaveBatch(urls map[string]string) error
	Load(id string) (string, bool)
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
