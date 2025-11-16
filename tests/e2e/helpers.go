package e2e

import (
	"avito-autumn-2025/internal/config"
	"avito-autumn-2025/internal/http/server"
	"avito-autumn-2025/internal/logger"
	db "avito-autumn-2025/internal/postgres"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

var testServer *server.Server
var testDB *pgxpool.Pool

func SetupE2ETest(t *testing.T) {
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

	testDB, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}

	_, _ = testDB.Exec(context.Background(), "TRUNCATE TABLE pull_request_reviewers CASCADE")
	_, _ = testDB.Exec(context.Background(), "TRUNCATE TABLE pull_requests CASCADE")
	_, _ = testDB.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	_, _ = testDB.Exec(context.Background(), "TRUNCATE TABLE teams CASCADE")

	migrationsPath := filepath.Join("..", "..", "migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = filepath.Join(".", "migrations")
	}
	err = db.RunMigrations(testDB, migrationsPath)
	require.NoError(t, err)

	logger := logger.NewStdLogger()
	testServer = server.NewServer(testDB, logger)
	testServer.SetupRoutes()
}

func TeardownE2ETest() {
	if testDB != nil {
		testDB.Close()
	}
}

func GetTestServer() *server.Server {
	return testServer
}
