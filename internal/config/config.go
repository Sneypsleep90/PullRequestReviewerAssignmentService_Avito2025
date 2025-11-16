package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Host        string `env:"HOST" yaml:"host" env-default:"0.0.0.0"`
	Port        int    `env:"PORT" yaml:"port" env-default:"8080"`
	DatabaseURL string `env:"DATABASE_URL" yaml:"database_url"`

	DBHost     string `env:"DB_HOST" yaml:"db_host" env-default:"localhost"`
	DBPort     int    `env:"DB_PORT" yaml:"db_port" env-default:"5432"`
	DBUser     string `env:"DB_USER" yaml:"db_user" env-default:"myuser"`
	DBPassword string `env:"DB_PASSWORD" yaml:"db_password" env-default:"mypassword"`
	DBName     string `env:"DB_NAME" yaml:"db_name" env-default:"mydatabase"`
	DBSSLMode  string `env:"DB_SSLMODE" yaml:"db_sslmode" env-default:"disable"`

	TestDBName string `env:"TEST_DB_NAME" yaml:"test_db_name" env-default:"test_mydatabase"`
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

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		if err := cleanenv.ReadEnv(cfg); err != nil {
			return nil, fmt.Errorf("cannot read config: %w", err)
		}
	}
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = cfg.BuildDatabaseURL()
	}
	return cfg, nil
}
