package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/delyke/urlShortener/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

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

type ConflictError struct {
	ShortURL string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("url already exists with short URL: %s", e.ShortURL)
}

func NewConflictError(shortURL string) error {
	return &ConflictError{ShortURL: shortURL}
}

func (repo *PostgresRepository) Save(originalURL string, shortedURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	query := `
        INSERT INTO urls (original_url, short_url)
        VALUES ($1, $2)
        ON CONFLICT (original_url) DO NOTHING
    `
	result, err := repo.db.ExecContext(ctx, query, originalURL, shortedURL)
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

func (repo *PostgresRepository) GetOriginalLink(shortedURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var originalURL string
	err := repo.db.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url = $1", shortedURL).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrRecordNotFound
		} else {
			return "", err
		}
	}
	return originalURL, nil
}

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

func (repo *PostgresRepository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := repo.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
