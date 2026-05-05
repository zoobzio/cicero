package stores

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/cicero/models"
)

// Translations provides database access for translation records.
type Translations struct {
	*sum.Database[models.Translation]
}

// NewTranslations creates a new translations store.
func NewTranslations(db *sqlx.DB, renderer astql.Renderer) *Translations {
	return &Translations{Database: sum.NewDatabase[models.Translation](db, "translations", renderer)}
}

// GetBySourceAndLang retrieves a translation by source hash and language pair.
func (s *Translations) GetBySourceAndLang(ctx context.Context, sourceHash, sourceLang, targetLang string) (*models.Translation, error) {
	return s.Select().
		Where("source_hash", "=", "source_hash").
		Where("source_lang", "=", "source_lang").
		Where("target_lang", "=", "target_lang").
		Exec(ctx, map[string]any{
			"source_hash": sourceHash,
			"source_lang": sourceLang,
			"target_lang": targetLang,
		})
}

// ListBySourceHash retrieves all translations for a given source hash.
func (s *Translations) ListBySourceHash(ctx context.Context, sourceHash string) ([]*models.Translation, error) {
	return s.Query().
		Where("source_hash", "=", "source_hash").
		Exec(ctx, map[string]any{"source_hash": sourceHash})
}
