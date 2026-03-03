package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobzio/astql"
)

// Stores aggregates all data access implementations.
type Stores struct {
	Sources      *Sources
	Translations *Translations
}

// New creates all stores and returns the aggregate.
func New(db *sqlx.DB, renderer astql.Renderer) (*Stores, error) {
	sources, err := NewSources(db, renderer)
	if err != nil {
		return nil, err
	}

	translations, err := NewTranslations(db, renderer)
	if err != nil {
		return nil, err
	}

	return &Stores{
		Sources:      sources,
		Translations: translations,
	}, nil
}
