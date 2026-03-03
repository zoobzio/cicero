//go:build testing

package wire

import "testing"

func TestTranslateRequest_Validate(t *testing.T) {
	valid := TranslateRequest{
		Text:       "Hello, world!",
		SourceLang: "en",
		TargetLang: "es",
		TenantID:   "zoobzio",
	}

	tests := []struct {
		name    string
		mutate  func(*TranslateRequest)
		wantErr bool
	}{
		{
			name:    "valid",
			mutate:  nil,
			wantErr: false,
		},
		{
			name:    "missing text",
			mutate:  func(r *TranslateRequest) { r.Text = "" },
			wantErr: true,
		},
		{
			name:    "missing source_lang",
			mutate:  func(r *TranslateRequest) { r.SourceLang = "" },
			wantErr: true,
		},
		{
			name:    "missing target_lang",
			mutate:  func(r *TranslateRequest) { r.TargetLang = "" },
			wantErr: true,
		},
		{
			name:    "missing tenant_id",
			mutate:  func(r *TranslateRequest) { r.TenantID = "" },
			wantErr: true,
		},
		{
			name:    "source_lang equals target_lang",
			mutate:  func(r *TranslateRequest) { r.TargetLang = r.SourceLang },
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := valid
			if tc.mutate != nil {
				tc.mutate(&r)
			}
			err := r.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTranslationsByHashResponse_Clone_Independence(t *testing.T) {
	original := TranslationsByHashResponse{
		Hash:       "abc",
		SourceText: "Hello",
		Translations: []TranslationDetail{
			{SourceLang: "en", TargetLang: "es", TranslatedText: "Hola"},
		},
	}

	clone := original.Clone()

	// Mutating clone's slice does not affect original.
	clone.Translations[0].TranslatedText = "mutated"
	if original.Translations[0].TranslatedText == "mutated" {
		t.Error("mutating clone Translations affected original")
	}
}

func TestTranslationsByHashResponse_Clone_NilTranslations(t *testing.T) {
	original := TranslationsByHashResponse{Hash: "abc", Translations: nil}
	clone := original.Clone()
	if clone.Translations != nil {
		t.Error("clone.Translations should be nil when original is nil")
	}
}
