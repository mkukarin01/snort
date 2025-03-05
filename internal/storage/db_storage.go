package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
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

	// создаем таблицу с уникальными полями short_id и original_url
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_id VARCHAR(8) UNIQUE NOT NULL,
			original_url TEXT UNIQUE NOT NULL
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
		return ErrDBConnection
	}

	return d.db.Ping()
}

// Close - закрываем соединения бд
func (d *Database) Close() error {
	// для нул базы - вернем ошибку
	if d == nil || d.db == nil {
		// эта ошибка специфична только для подключения к реальной базе
		return errors.New("database connection is already closed or uninitialized")
	}

	return d.db.Close()
}

// Save - сохраняем f5, обрабатывая конфликты
func (d *Database) Save(id, url string) error {
	if d == nil || d.db == nil {
		return errors.New("database connection is nil")
	}

	_, err := d.db.Exec(`
		INSERT INTO urls (short_id, original_url) 
		VALUES ($1, $2)
	`, id, url)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case pgerrcode.UniqueViolation:
				if pqErr.Constraint == "urls_original_url_key" { // unique для original_url
					return ErrURLConflict
				}
				if pqErr.Constraint == "urls_short_id_key" { // unique для short_id
					return ErrShortIDConflict
				}
			}
		}
		return err
	}

	return nil
}

// SaveBatch - сохраняемся несколько раз через комит - f5 + f5 + f5
func (d *Database) SaveBatch(urls map[string]string) error {
	if d == nil || d.db == nil {
		return errors.New("database connection is nil")
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO urls (short_id, original_url) 
		VALUES ($1, $2) 
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for id, url := range urls {
		_, err := stmt.Exec(id, url)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// Load - загружаеся f8
func (d *Database) Load(id string) (string, error) {
	// для нул базы - ничего не возвращаем
	if d == nil || d.db == nil {
		return "", ErrDBConnection
	}

	var url string
	err := d.db.QueryRow("SELECT original_url FROM urls WHERE short_id = $1", id).Scan(&url)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrURLNotFound
	}
	if err != nil {
		return "", err
	}

	return url, nil
}

// FindIDByURL находит short_id по original_url
func (d *Database) FindIDByURL(url string) (string, error) {
	if d == nil || d.db == nil {
		return "", ErrDBConnection
	}

	var shortID string
	err := d.db.QueryRow("SELECT short_id FROM urls WHERE original_url = $1", url).Scan(&shortID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrURLNotFound
	}
	if err != nil {
		return "", err
	}

	return shortID, nil
}
