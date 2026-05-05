// Package classify provides text complexity classification for routing decisions.
package classify

import (
	"context"

	"github.com/zoobz-io/cicero/models"
)

// Classifier determines how text should be routed for translation.
type Classifier interface {
	Classify(ctx context.Context, text string) (models.Classification, error)
}

// Simple always routes to the sidecar translator.
// This implementation is used in the first slice; the full deterministic
// algorithm with markdown detection and structural analysis lands alongside
// the LLM provider in a subsequent issue.
type Simple struct{}

// Classify returns a simple classification that routes to the sidecar.
func (s *Simple) Classify(_ context.Context, _ string) (models.Classification, error) {
	return models.Classification{Route: models.RouteSimple}, nil
}
