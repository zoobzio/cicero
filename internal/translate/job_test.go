//go:build testing

package translate

import (
	"testing"

	"github.com/zoobz-io/cicero/models"
)

func TestJob_Clone_Independence(t *testing.T) {
	existing := &models.Translation{
		ID:     1,
		Status: "completed",
		Text:   "¡Hola, mundo!",
	}
	original := &Job{
		SourceText:     "Hello, world!",
		SourceLang:     "en",
		TargetLang:     "es",
		TenantID:       "zoobzio",
		Hash:           "34ee2e3c1d6d112eab804965da0388e9",
		TranslatedText: "¡Hola, mundo!",
		Provider:       "sidecar",
		Status:         "completed",
		Existing:       existing,
		Classification: models.Classification{
			Route:   models.RouteSimple,
			Signals: []string{"length", "structure"},
		},
	}

	clone := original.Clone()

	// All scalar fields copied correctly.
	if clone.SourceText != original.SourceText {
		t.Errorf("SourceText: got %q, want %q", clone.SourceText, original.SourceText)
	}
	if clone.Hash != original.Hash {
		t.Errorf("Hash: got %q, want %q", clone.Hash, original.Hash)
	}
	if clone.Status != original.Status {
		t.Errorf("Status: got %q, want %q", clone.Status, original.Status)
	}
	if clone.Classification.Route != original.Classification.Route {
		t.Errorf("Classification.Route: got %q, want %q", clone.Classification.Route, original.Classification.Route)
	}

	// Classification.Signals is a deep copy — mutating clone does not affect original.
	if len(clone.Classification.Signals) != len(original.Classification.Signals) {
		t.Errorf("Signals length: got %d, want %d", len(clone.Classification.Signals), len(original.Classification.Signals))
	}
	clone.Classification.Signals[0] = "mutated"
	if original.Classification.Signals[0] == "mutated" {
		t.Error("mutating clone Signals affected original")
	}

	// Existing is a deep copy — mutating clone does not affect original.
	if clone.Existing == nil {
		t.Fatal("clone.Existing is nil")
	}
	clone.Existing.Status = "failed"
	if original.Existing.Status == "failed" {
		t.Error("mutating clone.Existing affected original.Existing")
	}

	// Clone pointer is independent of original pointer.
	if clone.Existing == original.Existing {
		t.Error("clone.Existing points to same struct as original.Existing")
	}
}

func TestJob_Clone_NilExisting(t *testing.T) {
	original := &Job{
		SourceText: "Hello",
		SourceLang: "en",
		TargetLang: "es",
	}

	clone := original.Clone()

	if clone.Existing != nil {
		t.Error("clone.Existing should be nil when original.Existing is nil")
	}
}

func TestJob_Clone_NilSignals(t *testing.T) {
	original := &Job{
		Classification: models.Classification{
			Route:   models.RouteSimple,
			Signals: nil,
		},
	}

	clone := original.Clone()

	if clone.Classification.Signals != nil {
		t.Error("clone.Classification.Signals should be nil when original is nil")
	}
}
