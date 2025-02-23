package storage

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq" // какой-то драйвер
)

// Storager - интерфейс для работы с бд или другим хранилищем
// (ага, можно же как-то потом подключится к ./storage.json)
type Storager interface {
	Ping() error
	Close() error
}

// Database обертка вокруг *sql.DB, реализует Storager
type Database struct {
	db *sql.DB
}

// NewDatabase запускатор соединения с pg или БД
func NewDatabase(dsn string) (*Database, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// пинганем соединение
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db: db}, nil
}

// Ping - проверка соединения с базой данных
func (d *Database) Ping() error {
	if d == nil || d.db == nil {
		return errors.New("database connection is nil")
	}
	return d.db.Ping()
}

func (d *Database) Close() error {
	if d == nil || d.db == nil {
		return errors.New("database connection is already closed or uninitialized")
	}
	return d.db.Close()
}
