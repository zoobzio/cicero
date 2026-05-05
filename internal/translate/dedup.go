package translate

import (
	"context"
	"errors"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/sum"
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
