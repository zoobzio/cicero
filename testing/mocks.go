//go:build testing

package testing

import (
	"context"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/cicero/internal/classify"
	"github.com/zoobz-io/cicero/internal/translate"
	"github.com/zoobz-io/cicero/models"
	"github.com/zoobz-io/pipz"
)

// Compile-time assertions: mocks satisfy their contracts.
var _ contracts.Sources = (*MockSources)(nil)
var _ contracts.Translations = (*MockTranslations)(nil)
var _ contracts.Translator = (*MockTranslator)(nil)
var _ classify.Classifier = (*MockClassifier)(nil)

// MockSources is a mock implementation of api/contracts.Sources.
type MockSources struct {
	OnGet func(ctx context.Context, hash string) (*models.Source, error)
	OnSet func(ctx context.Context, hash string, source *models.Source) error
}

func (m *MockSources) Get(ctx context.Context, hash string) (*models.Source, error) {
	if m.OnGet != nil {
		return m.OnGet(ctx, hash)
	}
	return &models.Source{}, nil
}

func (m *MockSources) Set(ctx context.Context, hash string, source *models.Source) error {
	if m.OnSet != nil {
		return m.OnSet(ctx, hash, source)
	}
	return nil
}

// MockTranslations is a mock implementation of api/contracts.Translations.
type MockTranslations struct {
	OnGetBySourceAndLang func(ctx context.Context, sourceHash, sourceLang, targetLang string) (*models.Translation, error)
	OnListBySourceHash   func(ctx context.Context, sourceHash string) ([]*models.Translation, error)
	OnSet                func(ctx context.Context, key string, translation *models.Translation) error
}

func (m *MockTranslations) GetBySourceAndLang(ctx context.Context, sourceHash, sourceLang, targetLang string) (*models.Translation, error) {
	if m.OnGetBySourceAndLang != nil {
		return m.OnGetBySourceAndLang(ctx, sourceHash, sourceLang, targetLang)
	}
	return nil, nil
}

func (m *MockTranslations) ListBySourceHash(ctx context.Context, sourceHash string) ([]*models.Translation, error) {
	if m.OnListBySourceHash != nil {
		return m.OnListBySourceHash(ctx, sourceHash)
	}
	return nil, nil
}

func (m *MockTranslations) Set(ctx context.Context, key string, translation *models.Translation) error {
	if m.OnSet != nil {
		return m.OnSet(ctx, key, translation)
	}
	return nil
}

// MockTranslator is a mock implementation of api/contracts.Translator.
type MockTranslator struct {
	OnTranslate func(ctx context.Context, text, sourceLang, targetLang string, route models.Route) (string, string, error)
}

func (m *MockTranslator) Translate(ctx context.Context, text, sourceLang, targetLang string, route models.Route) (string, string, error) {
	if m.OnTranslate != nil {
		return m.OnTranslate(ctx, text, sourceLang, targetLang, route)
	}
	return "", "", nil
}

// MockClassifier is a mock implementation of classify.Classifier.
type MockClassifier struct {
	OnClassify func(ctx context.Context, text string) (models.Classification, error)
}

func (m *MockClassifier) Classify(ctx context.Context, text string) (models.Classification, error) {
	if m.OnClassify != nil {
		return m.OnClassify(ctx, text)
	}
	return models.Classification{Route: models.RouteSimple}, nil
}

// MockPipeline is a mock implementation of pipz.Chainable[*translate.Job].
type MockPipeline struct {
	OnProcess func(ctx context.Context, job *translate.Job) (*translate.Job, error)
}

func (m *MockPipeline) Process(ctx context.Context, job *translate.Job) (*translate.Job, error) {
	if m.OnProcess != nil {
		return m.OnProcess(ctx, job)
	}
	return job, nil
}

func (m *MockPipeline) Identity() pipz.Identity {
	return pipz.NewIdentity("mock.pipeline", "Mock translation pipeline")
}

func (m *MockPipeline) Schema() pipz.Node {
	return pipz.Node{}
}

func (m *MockPipeline) Close() error {
	return nil
}

// Compile-time assertion: MockPipeline satisfies pipz.Chainable[*translate.Job].
var _ pipz.Chainable[*translate.Job] = (*MockPipeline)(nil)
