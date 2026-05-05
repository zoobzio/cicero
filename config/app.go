// Package config provides configuration types for cicero.
package config

import "github.com/zoobz-io/check"

// App holds application server configuration.
type App struct {
	Port int `env:"APP_PORT" default:"8080"`
}

// Validate checks that the app configuration is valid.
func (a App) Validate() error {
	return check.All(
		check.Int(a.Port, "port").Positive().Max(65535).V(),
	).Err()
}
