package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Host        string `env:"HOST" env-default:"0.0.0.0"`
	Port        int    `env:"PORT" env-default:"8080"`
	DatabaseURL string `env:"DATABASE_URL"`

	DBHost     string `env:"DB_HOST" env-default:"localhost"`
	DBPort     int    `env:"DB_PORT" env-default:"5432"`
	DBUser     string `env:"DB_USER" env-default:"myuser"`
	DBPassword string `env:"DB_PASSWORD" env-default:"mypassword"`
	DBName     string `env:"DB_NAME" env-default:"mydatabase"`
	DBSSLMode  string `env:"DB_SSLMODE" env-default:"disable"`

	TestDBName string `env:"TEST_DB_NAME" env-default:"test_mydatabase"`
}

func (c *Config) BuildDatabaseURL() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func (c *Config) BuildTestDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.TestDBName, c.DBSSLMode)
}

func (c *Config) BuildAdminDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBSSLMode)
}

func LoadConfig(envFiles ...string) (*Config, error) {
	paths := make([]string, 0, len(envFiles))
	for _, path := range envFiles {
		if path != "" {
			paths = append(paths, path)
		}
	}
	if len(paths) == 0 {
		paths = append(paths, ".env")
	}

	for _, path := range paths {
		if err := godotenv.Overload(path); err != nil {
			var pathErr *os.PathError
			if errors.As(err, &pathErr) && errors.Is(pathErr, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("cannot load env file %s: %w", path, err)
		}
	}

	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("cannot read env config: %w", err)
	}
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = cfg.BuildDatabaseURL()
	}
	return cfg, nil
}
