package contracts

import (
	"context"

	"github.com/zoobz-io/cicero/models"
)

// Translations defines the contract for translation storage operations.
type Translations interface {
	// GetBySourceAndLang retrieves a translation by source hash and language pair.
	GetBySourceAndLang(ctx context.Context, sourceHash, sourceLang, targetLang string) (*models.Translation, error)
	// ListBySourceHash retrieves all translations associated with a source hash.
	ListBySourceHash(ctx context.Context, sourceHash string) ([]*models.Translation, error)
	// Set stores a translation record, inserting or updating by primary key.
	Set(ctx context.Context, key string, translation *models.Translation) error
}
