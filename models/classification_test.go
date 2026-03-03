//go:build testing

package models

import "testing"

func TestRoute_Constants(t *testing.T) {
	if RouteSimple != Route("simple") {
		t.Errorf("RouteSimple: got %q, want %q", RouteSimple, "simple")
	}
	if RouteComplex != Route("complex") {
		t.Errorf("RouteComplex: got %q, want %q", RouteComplex, "complex")
	}
}

func TestClassification_ZeroValue(t *testing.T) {
	var c Classification
	if c.Route != "" {
		t.Errorf("zero Route: got %q, want empty string", c.Route)
	}
	if c.Signals != nil {
		t.Errorf("zero Signals: got %v, want nil", c.Signals)
	}
}

func TestClassification_WithSignals(t *testing.T) {
	c := Classification{
		Route:   RouteSimple,
		Signals: []string{"length", "structure"},
	}

	if c.Route != RouteSimple {
		t.Errorf("Route: got %q, want %q", c.Route, RouteSimple)
	}
	if len(c.Signals) != 2 {
		t.Errorf("Signals length: got %d, want 2", len(c.Signals))
	}
}
