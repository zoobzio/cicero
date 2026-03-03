// Package stores provides shared data access layer implementations for cicero.
package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobzio/astql"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/cicero/models"
)

// Sources provides database access for source text records.
// The source hash is the primary key; basic CRUD via sum.Database is sufficient.
type Sources struct {
	*sum.Database[models.Source]
}

// NewSources creates a new sources store.
func NewSources(db *sqlx.DB, renderer astql.Renderer) (*Sources, error) {
	database, err := sum.NewDatabase[models.Source](db, "sources", renderer)
	if err != nil {
		return nil, err
	}
	return &Sources{Database: database}, nil
}
