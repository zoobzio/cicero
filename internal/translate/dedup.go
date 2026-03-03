package translate

import (
	"context"
	"errors"

	"github.com/zoobzio/cicero/api/contracts"
	"github.com/zoobzio/grub"
	"github.com/zoobzio/sum"
)

// deduplicateStage checks whether a translation already exists for the
// given source hash and language pair. If found, the existing translation
// is stored in the job and downstream stages (translate, store) will skip.
func deduplicateStage(ctx context.Context, job *Job) (*Job, error) {
	translations := sum.MustUse[contracts.Translations](ctx)

	existing, err := translations.GetBySourceAndLang(ctx, job.Hash, job.SourceLang, job.TargetLang)
	if err != nil {
		if errors.Is(err, grub.ErrNotFound) {
			return job, nil
		}
		return job, err
	}
	if existing == nil {
		return job, nil
	}

	job.Existing = existing
	job.TranslatedText = existing.Text
	job.Provider = existing.Provider
	job.Status = existing.Status
	return job, nil
}
