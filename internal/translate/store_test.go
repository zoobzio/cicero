//go:build testing

package translate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zoobzio/cicero/api/contracts"
	"github.com/zoobzio/cicero/models"
	"github.com/zoobzio/sum"
)

func TestStoreStage_PersistsSourceAndTranslation(t *testing.T) {
	var storedSource *models.Source
	var storedTranslation *models.Translation

	ms := &mockSources{
		set: func(_ context.Context, _ string, source *models.Source) error {
			storedSource = source
			return nil
		},
	}
	mt := &mockTranslations{
		set: func(_ context.Context, _ string, translation *models.Translation) error {
			storedTranslation = translation
			return nil
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Sources](k, ms)
	sum.Register[contracts.Translations](k, mt)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{
		Hash:           "34ee2e3c1d6d112eab804965da0388e9",
		SourceText:     "Hello, world!",
		SourceLang:     "en",
		TargetLang:     "es",
		TenantID:       "zoobzio",
		TranslatedText: "¡Hola, mundo!",
		Provider:       "sidecar",
		Status:         "completed",
	}

	_, err := storeStage(ctx, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if storedSource == nil {
		t.Fatal("source was not persisted")
	}
	if storedSource.Hash != job.Hash {
		t.Errorf("source Hash: got %q, want %q", storedSource.Hash, job.Hash)
	}
	if storedSource.Text != job.SourceText {
		t.Errorf("source Text: got %q, want %q", storedSource.Text, job.SourceText)
	}
	if storedSource.TenantID != job.TenantID {
		t.Errorf("source TenantID: got %q, want %q", storedSource.TenantID, job.TenantID)
	}

	if storedTranslation == nil {
		t.Fatal("translation was not persisted")
	}
	if storedTranslation.SourceHash != job.Hash {
		t.Errorf("translation SourceHash: got %q, want %q", storedTranslation.SourceHash, job.Hash)
	}
	if storedTranslation.Text != job.TranslatedText {
		t.Errorf("translation Text: got %q, want %q", storedTranslation.Text, job.TranslatedText)
	}
	if storedTranslation.Provider != job.Provider {
		t.Errorf("translation Provider: got %q, want %q", storedTranslation.Provider, job.Provider)
	}
	if storedTranslation.Status != job.Status {
		t.Errorf("translation Status: got %q, want %q", storedTranslation.Status, job.Status)
	}
}

func TestStoreStage_SkipsWhenExistingSet(t *testing.T) {
	sourceCalls := 0
	translationCalls := 0

	ms := &mockSources{
		set: func(_ context.Context, _ string, _ *models.Source) error {
			sourceCalls++
			return nil
		},
	}
	mt := &mockTranslations{
		set: func(_ context.Context, _ string, _ *models.Translation) error {
			translationCalls++
			return nil
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Sources](k, ms)
	sum.Register[contracts.Translations](k, mt)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{
		Existing: &models.Translation{ID: 1, Status: "completed"},
	}

	_, err := storeStage(ctx, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sourceCalls != 0 {
		t.Errorf("sources.Set called %d times, want 0", sourceCalls)
	}
	if translationCalls != 0 {
		t.Errorf("translations.Set called %d times, want 0", translationCalls)
	}
}

func TestStoreStage_SourceError_Propagates(t *testing.T) {
	sourceErr := errors.New("source write failed")

	ms := &mockSources{
		set: func(_ context.Context, _ string, _ *models.Source) error {
			return sourceErr
		},
	}
	mt := &mockTranslations{}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Sources](k, ms)
	sum.Register[contracts.Translations](k, mt)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{
		Hash:       "hash",
		SourceText: "Hello",
		TenantID:   "zoobzio",
		Status:     "completed",
	}

	_, err := storeStage(ctx, job)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sourceErr) {
		t.Errorf("error: got %v, want to wrap %v", err, sourceErr)
	}
}

func TestStoreStage_TranslationError_Propagates(t *testing.T) {
	translationErr := errors.New("translation write failed")

	ms := &mockSources{}
	mt := &mockTranslations{
		set: func(_ context.Context, _ string, _ *models.Translation) error {
			return translationErr
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Sources](k, ms)
	sum.Register[contracts.Translations](k, mt)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{
		Hash:       "hash",
		SourceText: "Hello",
		TenantID:   "zoobzio",
		Status:     "completed",
	}

	_, err := storeStage(ctx, job)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, translationErr) {
		t.Errorf("error: got %v, want to wrap %v", err, translationErr)
	}
}
