package repository

import (
	"context"
	"database/sql"
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
	return &PostgresRepository{db: db}, nil
}

func (repo *PostgresRepository) Save(originalURL string, shortedURL string) error {
	return nil
}

func (repo *PostgresRepository) GetOriginalLink(shortedURL string) (string, error) {
	return "", nil
}

func (repo *PostgresRepository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := repo.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
