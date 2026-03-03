//go:build testing

package models

import (
	"testing"
	"time"
)

func TestSource_Clone(t *testing.T) {
	original := Source{
		Hash:      "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
		Text:      "Hello, world!",
		TenantID:  "zoobzio",
		CreatedAt: time.Now(),
	}

	clone := original.Clone()

	if clone.Hash != original.Hash {
		t.Errorf("Hash: got %q, want %q", clone.Hash, original.Hash)
	}
	if clone.Text != original.Text {
		t.Errorf("Text: got %q, want %q", clone.Text, original.Text)
	}
	if clone.TenantID != original.TenantID {
		t.Errorf("TenantID: got %q, want %q", clone.TenantID, original.TenantID)
	}

	// Mutations to clone do not affect original.
	clone.Hash = "modified"
	if original.Hash == "modified" {
		t.Error("mutating clone affected original Hash")
	}
}

func TestSource_Validate(t *testing.T) {
	tests := []struct {
		name    string
		source  Source
		wantErr bool
	}{
		{
			name: "valid",
			source: Source{
				Hash:     "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
				Text:     "Hello, world!",
				TenantID: "zoobzio",
			},
			wantErr: false,
		},
		{
			name: "missing hash",
			source: Source{
				Text:     "Hello, world!",
				TenantID: "zoobzio",
			},
			wantErr: true,
		},
		{
			name: "missing text",
			source: Source{
				Hash:     "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
				TenantID: "zoobzio",
			},
			wantErr: true,
		},
		{
			name: "missing tenant_id",
			source: Source{
				Hash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
				Text: "Hello, world!",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.source.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
