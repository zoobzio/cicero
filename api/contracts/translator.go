package contracts

import "context"

// Translator defines the contract for external translation provider calls.
type Translator interface {
	// Translate sends text to the translation provider and returns the translated string.
	Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error)
}
