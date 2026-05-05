// Package stores provides shared data access layer implementations for cicero.
package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/cicero/models"
)

// Sources provides database access for source text records.
// The source hash is the primary key; basic CRUD via sum.Database is sufficient.
type Sources struct {
	*sum.Database[models.Source]
}

// NewSources creates a new sources store.
func NewSources(db *sqlx.DB, renderer astql.Renderer) *Sources {
	return &Sources{Database: sum.NewDatabase[models.Source](db, "sources", renderer)}
}
