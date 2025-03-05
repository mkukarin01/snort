package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

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

	// создаем таблицу с уникальными полями short_id и original_url + user_id + is_deleted
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_id VARCHAR(8) UNIQUE NOT NULL,
			original_url TEXT UNIQUE NOT NULL,
			user_id TEXT NOT NULL,
			is_deleted BOOLEAN NOT NULL DEFAULT false
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Database{db: db}, nil
}

// -- методы из старого интерфейса --

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

// Save - старый метод сохранения, подкинем пустой uid
func (d *Database) Save(id, url string) error {
	return d.SaveUserURL("", id, url)
}

// SaveBatch - старый метод сохранения пачки, подкинем так же uid === ""
func (d *Database) SaveBatch(urls map[string]string) error {
	return d.SaveBatchUserURLs("", urls)
}

// Load - загружаем ссылку по short_id, проверяем флаг удаления
func (d *Database) Load(id string) (string, error) {
	if d == nil || d.db == nil {
		return "", ErrDBConnection
	}

	var (
		url       string
		isDeleted bool
	)

	err := d.db.QueryRow(`
		SELECT original_url, is_deleted
		FROM urls
		WHERE short_id = $1
	`, id).Scan(&url, &isDeleted)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrURLNotFound
	}
	if err != nil {
		return "", err
	}
	if isDeleted {
		return "", ErrURLDeleted
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

// -- методы с юид --
// NOTE: SaveUserURL и SaveBatchUserURLs - используют обычный лог, потому что мне лень доработать логгер

// SaveUserURL - сохраняемся с uid
func (d *Database) SaveUserURL(userID, shortID, originalURL string) error {
	if d == nil || d.db == nil {
		return errors.New("database connection is nil")
	}

	_, err := d.db.Exec(`
		INSERT INTO urls (short_id, original_url, user_id) 
		VALUES ($1, $2, $3)
	`, shortID, originalURL, userID)

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

		log.Printf("DB Error: can't insert shortID=%s, originalURL=%s, userID=%s: %v",
			shortID, originalURL, userID, err)
		return fmt.Errorf("failed to insert userURL: %w", err)
	}

	log.Printf("DB Info: successfully inserted shortID=%s, originalURL=%s, userID=%s",
		shortID, originalURL, userID)

	return nil
}

// SaveBatchUserURLs - сохраняем пачку с uid
func (d *Database) SaveBatchUserURLs(userID string, urls map[string]string) error {
	if d == nil || d.db == nil {
		return errors.New("database connection is nil")
	}

	tx, err := d.db.Begin()
	if err != nil {
		log.Printf("DB Error: can't begin transaction for userID=%s: %v", userID, err)
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO urls (short_id, original_url, user_id) 
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		tx.Rollback()
		log.Printf("DB Error: can't prepare statement for userID=%s: %v", userID, err)
		return err
	}
	defer stmt.Close()

	for shortID, originalURL := range urls {
		_, execErr := stmt.Exec(shortID, originalURL, userID)
		if execErr != nil {
			tx.Rollback()
			log.Printf("DB Error: can't insert shortID=%s, originalURL=%s, userID=%s: %v",
				shortID, originalURL, userID, execErr)
			return fmt.Errorf("failed batch insert: %w", execErr)
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		log.Printf("DB Error: can't commit transaction for userID=%s: %v", userID, commitErr)
		return commitErr
	}

	log.Printf("DB Info: successfully inserted batch of %d URLs for userID=%s", len(urls), userID)
	return nil
}

// GetUserURLs возвращает всё для заданного userID
func (d *Database) GetUserURLs(userID string) ([]UserURL, error) {
	if d == nil || d.db == nil {
		return nil, errors.New("database connection is nil")
	}

	rows, err := d.db.Query(`
		SELECT short_id, original_url
		FROM urls
		WHERE user_id = $1 AND is_deleted = false
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []UserURL
	for rows.Next() {
		var s, o string
		if err := rows.Scan(&s, &o); err != nil {
			return nil, err
		}
		result = append(result, UserURL{ShortURL: s, OriginalURL: o})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// MarkUserURLsDeleted - batch update для uid и списка shortIDs
func (d *Database) MarkUserURLsDeleted(userID string, shortIDs []string) error {
	if d == nil || d.db == nil {
		return errors.New("database connection is nil")
	}

	if len(shortIDs) == 0 {
		return nil
	}

	query := `
		UPDATE urls
		SET is_deleted = true
		WHERE user_id = $1
		  AND short_id = ANY($2)
	`
	_, err := d.db.Exec(query, userID, pq.StringArray(shortIDs))
	if err != nil {
		log.Printf("DB Error: MarkUserURLsDeleted userID=%s, shortIDs=%v, err=%v",
			userID, shortIDs, err)
		return err
	}

	return nil
}
