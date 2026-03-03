//go:build testing

package config

import "testing"

func TestApp_Validate(t *testing.T) {
	tests := []struct {
		name    string
		app     App
		wantErr bool
	}{
		{
			name:    "valid default port",
			app:     App{Port: 8080},
			wantErr: false,
		},
		{
			name:    "valid port 1",
			app:     App{Port: 1},
			wantErr: false,
		},
		{
			name:    "valid port max",
			app:     App{Port: 65535},
			wantErr: false,
		},
		{
			name:    "zero port",
			app:     App{Port: 0},
			wantErr: true,
		},
		{
			name:    "negative port",
			app:     App{Port: -1},
			wantErr: true,
		},
		{
			name:    "port above max",
			app:     App{Port: 65536},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.app.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
