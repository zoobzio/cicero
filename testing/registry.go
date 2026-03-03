//go:build testing

package testing

import (
	"context"
	"testing"
	"time"

	"github.com/zoobzio/cicero/api/contracts"
	"github.com/zoobzio/cicero/internal/classify"
	"github.com/zoobzio/cicero/internal/translate"
	"github.com/zoobzio/pipz"
	"github.com/zoobzio/sum"
)

// RegistryOption configures a test registry.
type RegistryOption func(k sum.Key)

// SetupRegistry initialises a sum registry with the given options, freezes it,
// and registers a cleanup function to reset it after the test.
func SetupRegistry(t *testing.T, opts ...RegistryOption) context.Context {
	t.Helper()
	sum.Reset()
	k := sum.Start()
	for _, opt := range opts {
		opt(k)
	}
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// WithSources registers a contracts.Sources implementation.
func WithSources(s contracts.Sources) RegistryOption {
	return func(k sum.Key) { sum.Register[contracts.Sources](k, s) }
}

// WithTranslations registers a contracts.Translations implementation.
func WithTranslations(tr contracts.Translations) RegistryOption {
	return func(k sum.Key) { sum.Register[contracts.Translations](k, tr) }
}

// WithTranslator registers a contracts.Translator implementation.
func WithTranslator(t contracts.Translator) RegistryOption {
	return func(k sum.Key) { sum.Register[contracts.Translator](k, t) }
}

// WithClassifier registers a classify.Classifier implementation.
func WithClassifier(c classify.Classifier) RegistryOption {
	return func(k sum.Key) { sum.Register[classify.Classifier](k, c) }
}

// WithPipeline registers a pipz.Chainable[*translate.Job] implementation.
func WithPipeline(p pipz.Chainable[*translate.Job]) RegistryOption {
	return func(k sum.Key) { sum.Register[pipz.Chainable[*translate.Job]](k, p) }
}
