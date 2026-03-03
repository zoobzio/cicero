package config

import (
	"fmt"

	"github.com/zoobzio/check"
)

// Database holds PostgreSQL connection configuration.
type Database struct {
	Host     string `env:"APP_DB_HOST" default:"localhost"`
	Port     int    `env:"APP_DB_PORT" default:"5432"`
	Name     string `env:"APP_DB_NAME" default:"cicero"`
	User     string `env:"APP_DB_USER" default:"cicero"`
	Password string `env:"APP_DB_PASSWORD" default:"cicero"`
	SSLMode  string `env:"APP_DB_SSLMODE" default:"disable"`
}

// DSN returns the PostgreSQL connection string.
func (d Database) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// Validate checks that the database configuration is valid.
func (d Database) Validate() error {
	return check.All(
		check.Str(d.Host, "host").Required().V(),
		check.Int(d.Port, "port").Positive().Max(65535).V(),
		check.Str(d.Name, "name").Required().V(),
		check.Str(d.User, "user").Required().V(),
		check.Str(d.SSLMode, "ssl_mode").Required().V(),
	).Err()
}
