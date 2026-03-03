package stores

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zoobzio/astql"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/cicero/models"
)

// Translations provides database access for translation records.
type Translations struct {
	*sum.Database[models.Translation]
}

// NewTranslations creates a new translations store.
func NewTranslations(db *sqlx.DB, renderer astql.Renderer) (*Translations, error) {
	database, err := sum.NewDatabase[models.Translation](db, "translations", renderer)
	if err != nil {
		return nil, err
	}
	return &Translations{Database: database}, nil
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
