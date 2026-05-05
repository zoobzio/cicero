package config

import "github.com/zoobz-io/check"

// Translator holds configuration for the translator gRPC sidecar client.
type Translator struct {
	Addr string `env:"APP_TRANSLATOR_ADDR" default:"localhost:9091"`
}

// Validate checks that the translator configuration is valid.
func (t Translator) Validate() error {
	return check.All(
		check.Str(t.Addr, "addr").Required().V(),
	).Err()
}
