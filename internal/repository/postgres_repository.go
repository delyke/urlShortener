package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"

	"github.com/delyke/urlShortener/internal/model"
)

// PostgresRepository stores URL data in a PostgreSQL database
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository connects to PostgreSQL and runs migrations.
func NewPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := RunMigrations(db, "migrations"); err != nil {
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

// ConflictError indicates a URL already exists with a short code.
type ConflictError struct {
	ShortURL string
}

// Error returns the conflict error string
func (e *ConflictError) Error() string {
	return fmt.Sprintf("url already exists with short URL: %s", e.ShortURL)
}

// NewConflictError - creates an error describing the conflicting short URL.
func NewConflictError(shortURL string) error {
	return &ConflictError{ShortURL: shortURL}
}

// DeleteURLsByUser - marks URLs as deleted for the specified user.
func (repo *PostgresRepository) DeleteURLsByUser(userID int64, URLs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND short_url = ANY($2)`
	result, err := repo.db.ExecContext(ctx, query, userID, pq.Array(URLs))
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	log.Printf("deleted %d urls, user ID: %d", n, userID)
	return nil
}

// GetURLsByUserID returns all URLs for the specified user.
func (repo *PostgresRepository) GetURLsByUserID(userID int64) (*[]model.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var urls []model.URL
	query := `SELECT * FROM urls WHERE user_id = $1`
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var url model.URL
		if err := rows.Scan(&url.UUID, &url.OriginalURL, &url.ShortURL, &url.UserID, &url.IsDeleted); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(urls) == 0 {
		return nil, ErrRecordNotFound
	}
	return &urls, nil
}

// Save stores a new URL mapping, returning a conflict error on duplicates.
func (repo *PostgresRepository) Save(originalURL string, shortedURL string, userID int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	query := `
        INSERT INTO urls (original_url, short_url, user_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (original_url) DO NOTHING
    `
	result, err := repo.db.ExecContext(ctx, query, originalURL, shortedURL, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			existing, err := repo.GetShortURLByOriginal(originalURL)
			if err != nil {
				return "", err
			}
			return "", NewConflictError(existing)
		}
		return "", err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", err
	}

	if rowsAffected == 0 {
		existing, err := repo.GetShortURLByOriginal(originalURL)
		if err != nil {
			return "", err
		}
		return existing, NewConflictError(existing)
	}

	return shortedURL, nil
}

// GetShortURLByOriginal - finds a short URL by its original URL.
func (repo *PostgresRepository) GetShortURLByOriginal(originalURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var shortedURL string
	err := repo.db.QueryRowContext(ctx,
		"SELECT short_url FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&shortedURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrRecordNotFound
		}
		return "", err
	}

	return shortedURL, nil
}

// GetOriginalLink - returns the original URL and deletion flag for a shortened URL.
func (repo *PostgresRepository) GetOriginalLink(shortedURL string) (string, *bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var originalURL string
	var isDeleted bool
	err := repo.db.QueryRowContext(ctx, "SELECT original_url, is_deleted FROM urls WHERE short_url = $1", shortedURL).Scan(&originalURL, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, ErrRecordNotFound
		} else {
			return "", nil, err
		}
	}
	return originalURL, &isDeleted, nil
}

// SaveBatch stores a batch of URL records in a transaction.
func (repo *PostgresRepository) SaveBatch(records []model.URL) error {
	tx, err := repo.db.Begin()
	if err != nil {
		log.Println("Begin error:", err)
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO urls (original_url, short_url) VALUES ($1, $2)")
	if err != nil {
		log.Println("Prepare error:", err)
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, record := range records {
		_, err := stmt.Exec(record.OriginalURL, record.ShortURL)
		if err != nil {
			log.Printf("Insert error for %s -> %s: %v", record.OriginalURL, record.ShortURL, err)
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println("Commit error:", err)
	}
	return nil
}

// Ping checks database connectivity
func (repo *PostgresRepository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := repo.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

// CreateUser inserts a new user record and returns its ID.
func (repo *PostgresRepository) CreateUser() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var id int64
	if err := repo.db.QueryRowContext(ctx, "INSERT INTO users DEFAULT VALUES RETURNING id").Scan(&id); err != nil {
		return -1, err
	}
	return id, nil
}
