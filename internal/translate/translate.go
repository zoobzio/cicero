package translate

import (
	"context"
	"fmt"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/sum"
)

// translateStage calls the translation provider to produce the translated text.
// If a cached translation already exists (set by deduplicateStage), this stage
// is a no-op. The classification route is forwarded to the provider so it can
// select the appropriate backend. On success, Provider is set from the provider
// response and Status to "completed". On error, the error is returned and pipz
// discards the carrier — the handler is responsible for determining "failed"
// status from the pipeline error.
func translateStage(ctx context.Context, job *Job) (*Job, error) {
	if job.Existing != nil {
		return job, nil
	}

	translator := sum.MustUse[contracts.Translator](ctx)

	translated, provider, err := translator.Translate(ctx, job.SourceText, job.SourceLang, job.TargetLang, job.Classification.Route)
	if err != nil {
		return job, fmt.Errorf("translate: %w", err)
	}

	job.TranslatedText = translated
	job.Provider = provider
	job.Status = "completed"
	return job, nil
}
