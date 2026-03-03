//go:build testing

package config

import (
	"strings"
	"testing"
)

func TestDatabase_DSN(t *testing.T) {
	db := Database{
		Host:     "localhost",
		Port:     5432,
		Name:     "cicero",
		User:     "cicero",
		Password: "secret",
		SSLMode:  "disable",
	}

	dsn := db.DSN()

	checks := []struct {
		name     string
		contains string
	}{
		{"host", "host=localhost"},
		{"port", "port=5432"},
		{"user", "user=cicero"},
		{"password", "password=secret"},
		{"dbname", "dbname=cicero"},
		{"sslmode", "sslmode=disable"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(dsn, c.contains) {
				t.Errorf("DSN %q does not contain %q", dsn, c.contains)
			}
		})
	}
}

func TestDatabase_DSN_PortInterpolation(t *testing.T) {
	db := Database{
		Host:    "db.example.com",
		Port:    5433,
		Name:    "mydb",
		User:    "admin",
		SSLMode: "require",
	}

	dsn := db.DSN()

	if !strings.Contains(dsn, "port=5433") {
		t.Errorf("DSN %q does not contain port=5433", dsn)
	}
	if !strings.Contains(dsn, "host=db.example.com") {
		t.Errorf("DSN %q does not contain host=db.example.com", dsn)
	}
}

func TestDatabase_Validate(t *testing.T) {
	valid := Database{
		Host:    "localhost",
		Port:    5432,
		Name:    "cicero",
		User:    "cicero",
		SSLMode: "disable",
	}

	tests := []struct {
		name    string
		mutate  func(*Database)
		wantErr bool
	}{
		{
			name:    "valid",
			mutate:  nil,
			wantErr: false,
		},
		{
			name:    "missing host",
			mutate:  func(d *Database) { d.Host = "" },
			wantErr: true,
		},
		{
			name:    "zero port",
			mutate:  func(d *Database) { d.Port = 0 },
			wantErr: true,
		},
		{
			name:    "port above max",
			mutate:  func(d *Database) { d.Port = 65536 },
			wantErr: true,
		},
		{
			name:    "missing name",
			mutate:  func(d *Database) { d.Name = "" },
			wantErr: true,
		},
		{
			name:    "missing user",
			mutate:  func(d *Database) { d.User = "" },
			wantErr: true,
		},
		{
			name:    "missing ssl_mode",
			mutate:  func(d *Database) { d.SSLMode = "" },
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := valid
			if tc.mutate != nil {
				tc.mutate(&db)
			}
			err := db.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
