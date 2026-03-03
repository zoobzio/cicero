//go:build testing

package config

import "testing"

func TestTranslator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Translator
		wantErr bool
	}{
		{
			name:    "valid",
			cfg:     Translator{Addr: "http://localhost:5000"},
			wantErr: false,
		},
		{
			name:    "missing addr",
			cfg:     Translator{Addr: ""},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
