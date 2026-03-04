//go:build testing

package contracts_test

import (
	"context"
	"testing"

	"github.com/zoobzio/cicero/api/contracts"
	"github.com/zoobzio/cicero/models"
	cicerotest "github.com/zoobzio/cicero/testing"
)

// Compile-time assertion: MockTranslator satisfies the updated Translator interface.
var _ contracts.Translator = (*cicerotest.MockTranslator)(nil)

func TestTranslatorInterface_RouteForwarded(t *testing.T) {
	var capturedRoute models.Route

	mock := &cicerotest.MockTranslator{
		OnTranslate: func(_ context.Context, _, _, _ string, route models.Route) (string, string, error) {
			capturedRoute = route
			return "¡Hola, mundo!", "sidecar", nil
		},
	}

	ctx := context.Background()
	result, provider, err := mock.Translate(ctx, "Hello", "en", "es", models.RouteSimple)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "¡Hola, mundo!" {
		t.Errorf("result: got %q, want %q", result, "¡Hola, mundo!")
	}
	if provider != "sidecar" {
		t.Errorf("provider: got %q, want %q", provider, "sidecar")
	}
	if capturedRoute != models.RouteSimple {
		t.Errorf("route: got %q, want %q", capturedRoute, models.RouteSimple)
	}
}

func TestTranslatorInterface_ZeroDefault(t *testing.T) {
	mock := &cicerotest.MockTranslator{}

	ctx := context.Background()
	result, provider, err := mock.Translate(ctx, "Hello", "en", "es", models.RouteSimple)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("zero result: got %q, want empty", result)
	}
	if provider != "" {
		t.Errorf("zero provider: got %q, want empty", provider)
	}
}

func TestTranslatorInterface_RouteComplex(t *testing.T) {
	var capturedRoute models.Route

	mock := &cicerotest.MockTranslator{
		OnTranslate: func(_ context.Context, _, _, _ string, route models.Route) (string, string, error) {
			capturedRoute = route
			return "Hola", "llm", nil
		},
	}

	ctx := context.Background()
	_, provider, err := mock.Translate(ctx, "Complex text", "en", "es", models.RouteComplex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedRoute != models.RouteComplex {
		t.Errorf("route: got %q, want %q", capturedRoute, models.RouteComplex)
	}
	if provider != "llm" {
		t.Errorf("provider: got %q, want %q", provider, "llm")
	}
}
