//go:build testing

package stores

import (
	"testing"

	"github.com/zoobz-io/cicero/api/contracts"
)

// Compile-time assertion: Translations satisfies contracts.Translations.
var _ contracts.Translations = (*Translations)(nil)

func TestTranslations_ImplementsContract(t *testing.T) {
	// Verified at compile time by the var _ assignment above.
	// This test documents that Translations satisfies contracts.Translations.
}
