//go:build testing

package classify

import (
	"context"
	"testing"

	"github.com/zoobz-io/cicero/models"
)

func TestSimple_Classify_AlwaysReturnsRouteSimple(t *testing.T) {
	s := &Simple{}
	ctx := context.Background()

	inputs := []string{
		"Hello, world!",
		"",
		"# Markdown heading\n\nWith **bold** content.",
		"A very long text that might otherwise trigger complexity escalation.",
	}

	for _, text := range inputs {
		result, err := s.Classify(ctx, text)
		if err != nil {
			t.Errorf("Classify(%q): unexpected error: %v", text, err)
		}
		if result.Route != models.RouteSimple {
			t.Errorf("Classify(%q): got route %q, want %q", text, result.Route, models.RouteSimple)
		}
		if result.Signals != nil {
			t.Errorf("Classify(%q): got signals %v, want nil", text, result.Signals)
		}
	}
}
