package storage

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq" // какой-то драйвер
)

// Database реализация хранилища в бд
type Database struct {
	db *sql.DB
}

// NewDatabase запускатор соединения с pg или БД
func NewDatabase(dsn string) (*Database, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN is empty")
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

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_id VARCHAR(8) UNIQUE NOT NULL,
			original_url TEXT NOT NULL
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Database{db: db}, nil
}

// Ping - проверка соединения с бд
func (d *Database) Ping() error {
	// для нул базы - вернем ошибку
	if d == nil || d.db == nil {
		return errors.New("database connection is nil")
	}

	return d.db.Ping()
}

// Close - закрываем соединения бд
func (d *Database) Close() error {
	// для нул базы - вернем ошибку
	if d == nil || d.db == nil {
		return errors.New("database connection is already closed or uninitialized")
	}

	return d.db.Close()
}

// Save - сохраняемся f5
func (d *Database) Save(id, url string) error {
	// для нул базы - вернем ошибку
	if d == nil || d.db == nil {
		return errors.New("database connection is nil")
	}

	_, err := d.db.Exec("INSERT INTO urls (short_id, original_url) VALUES ($1, $2) ON CONFLICT (short_id) DO NOTHING", id, url)
	return err
}

// Load - загружаеся f8
func (d *Database) Load(id string) (string, bool) {
	// для нул базы - ничего не возвращаем
	if d == nil || d.db == nil {
		return "", false
	}

	var url string
	err := d.db.QueryRow("SELECT original_url FROM urls WHERE short_id = $1", id).Scan(&url)
	if err != nil {
		return "", false
	}
	return url, true
}
