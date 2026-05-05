package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
)

// Stores aggregates all data access implementations.
type Stores struct {
	Sources      *Sources
	Translations *Translations
}

// New creates all stores and returns the aggregate.
func New(db *sqlx.DB, renderer astql.Renderer) *Stores {
	return &Stores{
		Sources:      NewSources(db, renderer),
		Translations: NewTranslations(db, renderer),
	}
}
