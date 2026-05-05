//go:build testing

package translate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zoobz-io/cicero/internal/classify"
	"github.com/zoobz-io/cicero/models"
	"github.com/zoobz-io/sum"
)

func TestClassifyStage_SetsClassification(t *testing.T) {
	want := models.Classification{
		Route:   models.RouteSimple,
		Signals: nil,
	}

	mc := &mockClassifier{
		classify: func(_ context.Context, _ string) (models.Classification, error) {
			return want, nil
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[classify.Classifier](k, mc)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{SourceText: "Hello, world!"}
	result, err := classifyStage(ctx, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Classification.Route != want.Route {
		t.Errorf("Route: got %q, want %q", result.Classification.Route, want.Route)
	}
}

func TestClassifyStage_ClassifierError_Propagates(t *testing.T) {
	classifyErr := errors.New("classifier failure")

	mc := &mockClassifier{
		classify: func(_ context.Context, _ string) (models.Classification, error) {
			return models.Classification{}, classifyErr
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[classify.Classifier](k, mc)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{SourceText: "Hello"}
	_, err := classifyStage(ctx, job)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, classifyErr) {
		t.Errorf("error: got %v, want to wrap %v", err, classifyErr)
	}
}

func TestClassifyStage_PassesSourceText(t *testing.T) {
	var capturedText string

	mc := &mockClassifier{
		classify: func(_ context.Context, text string) (models.Classification, error) {
			capturedText = text
			return models.Classification{Route: models.RouteSimple}, nil
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[classify.Classifier](k, mc)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	job := &Job{SourceText: "classify me"}
	_, _ = classifyStage(ctx, job)

	if capturedText != "classify me" {
		t.Errorf("classifier received %q, want %q", capturedText, "classify me")
	}
}
