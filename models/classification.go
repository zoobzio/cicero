package models

// Route represents the routing decision produced by the complexity classifier.
type Route string

const (
	// RouteSimple routes text to the self-hosted sidecar translator.
	RouteSimple Route = "simple"
	// RouteComplex routes text to an LLM for nuanced translation.
	RouteComplex Route = "complex"
)

// Classification holds the result of a complexity classification.
type Classification struct {
	Route   Route
	Signals []string // Signals that triggered this classification.
}
