//go:build testing

package stores

import (
	"testing"

	"github.com/zoobz-io/cicero/api/contracts"
)

// Compile-time assertion: Sources satisfies contracts.Sources.
var _ contracts.Sources = (*Sources)(nil)

func TestSources_ImplementsContract(t *testing.T) {
	// Verified at compile time by the var _ assignment above.
	// This test documents that Sources satisfies contracts.Sources.
}
