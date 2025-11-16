package integration

import (
	"avito-autumn-2025/internal/config"
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	cfg, err := config.LoadConfig("../../config.yaml")
	if err != nil {
		cfg, _ = config.LoadConfig("")
	}

	databaseURL := cfg.BuildTestDatabaseURL()
	adminURL := cfg.BuildAdminDatabaseURL()

	adminDB, err := sql.Open("pgx", adminURL)
	if err == nil {
		_, _ = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.TestDBName))
		adminDB.Close()
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil, nil
	}

	_, _ = pool.Exec(context.Background(), "TRUNCATE TABLE pull_request_reviewers CASCADE")
	_, _ = pool.Exec(context.Background(), "TRUNCATE TABLE pull_requests CASCADE")
	_, _ = pool.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	_, _ = pool.Exec(context.Background(), "TRUNCATE TABLE teams CASCADE")

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}
