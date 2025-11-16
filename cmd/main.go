package main

import (
	"avito-autumn-2025/internal/config"
	"avito-autumn-2025/internal/http/server"
	"avito-autumn-2025/internal/logger"
	db "avito-autumn-2025/internal/postgres"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		cfg, err = config.LoadConfig("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}
	}

	stdLogger := logger.NewStdLogger()
	stdLogger.Info("Starting reviewer service", "port", cfg.Port, "host", cfg.Host)

	dbPool, err := db.ConnectPostgres(cfg)
	if err != nil {
		stdLogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	migrationsPath := filepath.Join(".", "migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = filepath.Join("/root", "migrations")
		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			stdLogger.Error("Migrations directory not found", "tried_paths", []string{filepath.Join(".", "migrations"), migrationsPath})
			os.Exit(1)
		}
	}
	if err := db.RunMigrations(dbPool, migrationsPath); err != nil {
		stdLogger.Error("Failed to run migrations", "error", err, "migrations_path", migrationsPath)
		os.Exit(1)
	}
	stdLogger.Info("Migrations applied successfully", "migrations_path", migrationsPath)

	srv := server.NewServer(dbPool, stdLogger)
	srv.SetupRoutes()

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		stdLogger.Info("Starting HTTP server", "address", addr)
		if err := srv.Run(addr); err != nil {
			stdLogger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	stdLogger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		stdLogger.Error("Server forced to shutdown", "error", err)
	}

	stdLogger.Info("Server exited")
}
