//go:build testing

package translate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/cicero/models"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/sum"
)

func TestDedupStage_ExistingTranslation(t *testing.T) {
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

	mock := &mockTranslations{
		getBySourceAndLang: func(_ context.Context, _, _, _ string) (*models.Translation, error) {
			return existing, nil
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Translations](k, mock)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{
		Hash:       "34ee2e3c1d6d112eab804965da0388e9",
		SourceLang: "en",
		TargetLang: "es",
	}

	result, err := deduplicateStage(ctx, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Existing == nil {
		t.Fatal("job.Existing should be populated on cache hit")
	}
	if result.TranslatedText != existing.Text {
		t.Errorf("TranslatedText: got %q, want %q", result.TranslatedText, existing.Text)
	}
	if result.Provider != existing.Provider {
		t.Errorf("Provider: got %q, want %q", result.Provider, existing.Provider)
	}
	if result.Status != existing.Status {
		t.Errorf("Status: got %q, want %q", result.Status, existing.Status)
	}
}

func TestDedupStage_NotFound_GrubErrNotFound(t *testing.T) {
	mock := &mockTranslations{
		getBySourceAndLang: func(_ context.Context, _, _, _ string) (*models.Translation, error) {
			return nil, grub.ErrNotFound
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Translations](k, mock)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{Hash: "hash", SourceLang: "en", TargetLang: "es"}

	result, err := deduplicateStage(ctx, job)
	if err != nil {
		t.Fatalf("ErrNotFound should not propagate: %v", err)
	}
	if result.Existing != nil {
		t.Error("job.Existing should be nil when not found")
	}
}

func TestDedupStage_NotFound_NilResult(t *testing.T) {
	mock := &mockTranslations{
		getBySourceAndLang: func(_ context.Context, _, _, _ string) (*models.Translation, error) {
			return nil, nil
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Translations](k, mock)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{Hash: "hash", SourceLang: "en", TargetLang: "es"}

	result, err := deduplicateStage(ctx, job)
	if err != nil {
		t.Fatalf("nil result should not produce error: %v", err)
	}
	if result.Existing != nil {
		t.Error("job.Existing should be nil when store returns nil, nil")
	}
}

func TestDedupStage_StoreError_Propagates(t *testing.T) {
	storeErr := errors.New("database connection lost")

	mock := &mockTranslations{
		getBySourceAndLang: func(_ context.Context, _, _, _ string) (*models.Translation, error) {
			return nil, storeErr
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Translations](k, mock)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{Hash: "hash", SourceLang: "en", TargetLang: "es"}

	_, err := deduplicateStage(ctx, job)
	if err == nil {
		t.Fatal("expected error to propagate, got nil")
	}
	if !errors.Is(err, storeErr) {
		t.Errorf("error: got %v, want to wrap %v", err, storeErr)
	}
}

// mockTranslations is a local minimal mock used across translate package tests.
type mockTranslations struct {
	getBySourceAndLang func(ctx context.Context, sourceHash, sourceLang, targetLang string) (*models.Translation, error)
	listBySourceHash   func(ctx context.Context, sourceHash string) ([]*models.Translation, error)
	set                func(ctx context.Context, key string, translation *models.Translation) error
}

func (m *mockTranslations) GetBySourceAndLang(ctx context.Context, sourceHash, sourceLang, targetLang string) (*models.Translation, error) {
	if m.getBySourceAndLang != nil {
		return m.getBySourceAndLang(ctx, sourceHash, sourceLang, targetLang)
	}
	return nil, nil
}

func (m *mockTranslations) ListBySourceHash(ctx context.Context, sourceHash string) ([]*models.Translation, error) {
	if m.listBySourceHash != nil {
		return m.listBySourceHash(ctx, sourceHash)
	}
	return nil, nil
}

func (m *mockTranslations) Set(ctx context.Context, key string, translation *models.Translation) error {
	if m.set != nil {
		return m.set(ctx, key, translation)
	}
	return nil
}
