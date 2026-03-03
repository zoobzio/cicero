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

func TestTranslateStage_Success(t *testing.T) {
	mt := &mockTranslator{
		translate: func(_ context.Context, _, _, _ string) (string, error) {
			return "¡Hola, mundo!", nil
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Translator](k, mt)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{SourceText: "Hello, world!", SourceLang: "en", TargetLang: "es"}
	result, err := translateStage(ctx, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TranslatedText != "¡Hola, mundo!" {
		t.Errorf("TranslatedText: got %q, want %q", result.TranslatedText, "¡Hola, mundo!")
	}
	if result.Provider != "sidecar" {
		t.Errorf("Provider: got %q, want %q", result.Provider, "sidecar")
	}
	if result.Status != "completed" {
		t.Errorf("Status: got %q, want %q", result.Status, "completed")
	}
}

func TestTranslateStage_SkipsWhenExistingSet(t *testing.T) {
	callCount := 0
	mt := &mockTranslator{
		translate: func(_ context.Context, _, _, _ string) (string, error) {
			callCount++
			return "should not be called", nil
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Translator](k, mt)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{
		SourceText: "Hello",
		SourceLang: "en",
		TargetLang: "es",
		Existing: &models.Translation{
			Text:     "¡Hola!",
			Provider: "sidecar",
			Status:   "completed",
		},
		TranslatedText: "¡Hola!",
		Provider:       "sidecar",
		Status:         "completed",
	}

	result, err := translateStage(ctx, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 0 {
		t.Errorf("translator was called %d times, want 0", callCount)
	}
	// Output fields unchanged from dedup.
	if result.TranslatedText != "¡Hola!" {
		t.Errorf("TranslatedText: got %q, want %q", result.TranslatedText, "¡Hola!")
	}
}

func TestTranslateStage_TranslatorError_ReturnsError(t *testing.T) {
	translateErr := errors.New("sidecar unavailable")

	mt := &mockTranslator{
		translate: func(_ context.Context, _, _, _ string) (string, error) {
			return "", translateErr
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Translator](k, mt)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{SourceText: "Hello", SourceLang: "en", TargetLang: "es"}
	result, err := translateStage(ctx, job)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, translateErr) {
		t.Errorf("error: got %v, want to wrap %v", err, translateErr)
	}
	// translateStage does not set Status = "failed" on error.
	// The handler is responsible for setting failed status when the pipeline errors.
	if result.Status != "" {
		t.Errorf("Status: got %q, want empty string (handler sets failed status)", result.Status)
	}
}
