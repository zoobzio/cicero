package translate

import (
	"context"
	"fmt"

	"github.com/zoobzio/cicero/api/contracts"
	"github.com/zoobzio/cicero/models"
	"github.com/zoobzio/sum"
)

// storeStage persists the source text and translation records to the database.
// If a cached translation already exists (set by deduplicateStage), this stage
// is a no-op. The source record is idempotent via its hash primary key.
func storeStage(ctx context.Context, job *Job) (*Job, error) {
	if job.Existing != nil {
		return job, nil
	}

	sources := sum.MustUse[contracts.Sources](ctx)
	translations := sum.MustUse[contracts.Translations](ctx)

	source := &models.Source{
		Hash:     job.Hash,
		Text:     job.SourceText,
		TenantID: job.TenantID,
	}
	if err := sources.Set(ctx, job.Hash, source); err != nil {
		return job, fmt.Errorf("store source: %w", err)
	}

	translation := &models.Translation{
		SourceHash: job.Hash,
		SourceLang: job.SourceLang,
		TargetLang: job.TargetLang,
		Text:       job.TranslatedText,
		Provider:   job.Provider,
		Status:     job.Status,
		TenantID:   job.TenantID,
	}
	if err := translations.Set(ctx, "", translation); err != nil {
		return job, fmt.Errorf("store translation: %w", err)
	}

	return job, nil
}
