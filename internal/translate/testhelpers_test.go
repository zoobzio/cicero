//go:build testing

package translate

import (
	"context"

	"github.com/zoobzio/cicero/models"
)

// mockSources is a local minimal mock for contracts.Sources.
type mockSources struct {
	get func(ctx context.Context, hash string) (*models.Source, error)
	set func(ctx context.Context, hash string, source *models.Source) error
}

func (m *mockSources) Get(ctx context.Context, hash string) (*models.Source, error) {
	if m.get != nil {
		return m.get(ctx, hash)
	}
	return &models.Source{}, nil
}

func (m *mockSources) Set(ctx context.Context, hash string, source *models.Source) error {
	if m.set != nil {
		return m.set(ctx, hash, source)
	}
	return nil
}

// mockTranslator is a local minimal mock for contracts.Translator.
type mockTranslator struct {
	translate func(ctx context.Context, text, sourceLang, targetLang string, route models.Route) (string, string, error)
}

func (m *mockTranslator) Translate(ctx context.Context, text, sourceLang, targetLang string, route models.Route) (string, string, error) {
	if m.translate != nil {
		return m.translate(ctx, text, sourceLang, targetLang, route)
	}
	return "", "", nil
}

// mockClassifier is a local minimal mock for classify.Classifier.
type mockClassifier struct {
	classify func(ctx context.Context, text string) (models.Classification, error)
}

func (m *mockClassifier) Classify(ctx context.Context, text string) (models.Classification, error) {
	if m.classify != nil {
		return m.classify(ctx, text)
	}
	return models.Classification{Route: models.RouteSimple}, nil
}
