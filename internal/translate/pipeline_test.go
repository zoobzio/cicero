//go:build testing

package translate

import (
	"context"
	"testing"
	"time"

	"github.com/zoobzio/cicero/api/contracts"
	"github.com/zoobzio/cicero/internal/classify"
	"github.com/zoobzio/cicero/models"
	"github.com/zoobzio/sum"
)

func setupPipelineContext(t *testing.T, ms contracts.Sources, mt contracts.Translations, mtr contracts.Translator, mc classify.Classifier) context.Context {
	t.Helper()
	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Sources](k, ms)
	sum.Register[contracts.Translations](k, mt)
	sum.Register[contracts.Translator](k, mtr)
	sum.Register[classify.Classifier](k, mc)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestPipeline_NewTranslation_EndToEnd(t *testing.T) {
	ms := &mockSources{}
	mt := &mockTranslations{} // returns nil, nil — no existing translation
	mtr := &mockTranslator{
		translate: func(_ context.Context, _, _, _ string) (string, error) {
			return "¡Hola, mundo!", nil
		},
	}
	mc := &mockClassifier{}

	ctx := setupPipelineContext(t, ms, mt, mtr, mc)

	pipeline := NewPipeline()
	job := &Job{
		SourceText: "Hello, world!",
		SourceLang: "en",
		TargetLang: "es",
		TenantID:   "zoobzio",
	}

	result, err := pipeline.Process(ctx, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Hash stage ran.
	if len(result.Hash) != 32 {
		t.Errorf("Hash length: got %d, want 32", len(result.Hash))
	}
	const expectedHash = "315f5bdb76d078c43b8ac0064e4a0164"
	if result.Hash != expectedHash {
		t.Errorf("Hash: got %q, want %q", result.Hash, expectedHash)
	}

	// Classify stage ran.
	if result.Classification.Route != models.RouteSimple {
		t.Errorf("Classification.Route: got %q, want %q", result.Classification.Route, models.RouteSimple)
	}

	// Translate stage ran.
	if result.TranslatedText != "¡Hola, mundo!" {
		t.Errorf("TranslatedText: got %q, want %q", result.TranslatedText, "¡Hola, mundo!")
	}
	if result.Provider != "sidecar" {
		t.Errorf("Provider: got %q, want %q", result.Provider, "sidecar")
	}
	if result.Status != "completed" {
		t.Errorf("Status: got %q, want %q", result.Status, "completed")
	}

	// No existing translation.
	if result.Existing != nil {
		t.Error("Existing should be nil for a new translation")
	}
}

func TestPipeline_DedupHit_TranslatorNotCalled(t *testing.T) {
	existing := &models.Translation{
		ID:         1,
		SourceHash: "34ee2e3c1d6d112eab804965da0388e9",
		SourceLang: "en",
		TargetLang: "es",
		Text:       "¡Hola, mundo!",
		Provider:   "sidecar",
		Status:     "completed",
		TenantID:   "zoobzio",
	}

	ms := &mockSources{}
	mt := &mockTranslations{
		getBySourceAndLang: func(_ context.Context, _, _, _ string) (*models.Translation, error) {
			return existing, nil
		},
	}

	translatorCalled := false
	mtr := &mockTranslator{
		translate: func(_ context.Context, _, _, _ string) (string, error) {
			translatorCalled = true
			return "should not be returned", nil
		},
	}
	mc := &mockClassifier{}

	ctx := setupPipelineContext(t, ms, mt, mtr, mc)

	pipeline := NewPipeline()
	job := &Job{
		SourceText: "Hello, world!",
		SourceLang: "en",
		TargetLang: "es",
		TenantID:   "zoobzio",
	}

	result, err := pipeline.Process(ctx, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if translatorCalled {
		t.Error("translator should not be called on dedup hit")
	}
	if result.Existing == nil {
		t.Error("Existing should be set on dedup hit")
	}
	if result.TranslatedText != existing.Text {
		t.Errorf("TranslatedText: got %q, want %q", result.TranslatedText, existing.Text)
	}
	if result.Status != "completed" {
		t.Errorf("Status: got %q, want %q", result.Status, "completed")
	}
}

func TestPipeline_TranslatorError_PipelineFails(t *testing.T) {
	ms := &mockSources{}
	mt := &mockTranslations{} // no existing translation
	mtr := &mockTranslator{
		translate: func(_ context.Context, _, _, _ string) (string, error) {
			return "", context.DeadlineExceeded
		},
	}
	mc := &mockClassifier{}

	ctx := setupPipelineContext(t, ms, mt, mtr, mc)

	pipeline := NewPipeline()
	job := &Job{
		SourceText: "Hello",
		SourceLang: "en",
		TargetLang: "es",
		TenantID:   "zoobzio",
	}

	_, err := pipeline.Process(ctx, job)
	if err == nil {
		t.Fatal("expected error when translator fails, got nil")
	}
	// pipz.Apply returns zero (nil *Job) on error — result cannot be inspected.
	// The caller (handler) is responsible for setting Translation.Status = "failed".
}
