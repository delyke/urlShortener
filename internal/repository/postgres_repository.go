package repository

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
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

func (repo *PostgresRepository) Save(originalURL string, shortedURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := repo.db.ExecContext(ctx, "INSERT INTO urls (original_url, short_url) VALUES ($1, $2)", originalURL, shortedURL)
	if err != nil {
		return err
	}
	return nil
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

func (repo *PostgresRepository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := repo.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
