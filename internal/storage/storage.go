package storage

import (
	"github.com/mkukarin01/snort/internal/config"
)

// Storager - интерфейс для работы с бд или другим хранилищем
// (ага, можно же как-то потом подключится к ./storage.json)
type Storager interface {
	Ping() error
	Close() error
	Save(id, url string) error
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
