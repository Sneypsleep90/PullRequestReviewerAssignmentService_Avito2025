package db

import (
	"avito-autumn-2025/internal/config"
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"log"
	"time"
)

func ConnectPostgres(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	config := pool.Config()
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 2 * time.Minute

	log.Println("Successfully connected to database")
	return pool, nil
}

func RunMigrations(pool *pgxpool.Pool, migrationsDir string) error {
	connStr := pool.Config().ConnString()
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open sql.DB: %w", err)
	}
	defer sqlDB.Close()

	goose.SetDialect("postgres")

	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
