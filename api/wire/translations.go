// Package wire provides request and response types for the public API.
package wire

import "github.com/zoobzio/check"

// TranslateRequest is the request body for submitting text for translation.
type TranslateRequest struct {
	Text       string `json:"text" description:"Source text to translate" example:"Hello, world!"`
	SourceLang string `json:"source_lang" description:"Source language code" example:"en"`
	TargetLang string `json:"target_lang" description:"Target language code" example:"es"`
	TenantID   string `json:"tenant_id" description:"Tenant identifier" example:"zoobzio"`
}

// Validate validates the translation request.
func (r TranslateRequest) Validate() error {
	return check.All(
		check.Str(r.Text, "text").Required().V(),
		check.Str(r.SourceLang, "source_lang").Required().V(),
		check.Str(r.TargetLang, "target_lang").Required().V(),
		check.Str(r.TenantID, "tenant_id").Required().V(),
		check.NotEqual(r.SourceLang, r.TargetLang, "target_lang"),
	).Err()
}

// Clone returns a shallow copy.
func (r TranslateRequest) Clone() TranslateRequest {
	return r
}

// TranslateResponse is the response returned after submitting a translation.
type TranslateResponse struct {
	Hash           string `json:"hash" description:"Content hash for retrieval"`
	SourceText     string `json:"source_text" description:"Original text"`
	TranslatedText string `json:"translated_text" description:"Translated text"`
	SourceLang     string `json:"source_lang" description:"Source language code"`
	TargetLang     string `json:"target_lang" description:"Target language code"`
	Classification string `json:"classification" description:"Complexity classification" example:"simple"`
	Provider       string `json:"provider" description:"Translation provider used"`
	Status         string `json:"status" description:"Translation status"`
}

// Clone returns a shallow copy.
func (r TranslateResponse) Clone() TranslateResponse {
	return r
}

// TranslationsByHashResponse is the response for retrieving translations by content hash.
type TranslationsByHashResponse struct {
	Hash         string              `json:"hash"`
	SourceText   string              `json:"source_text"`
	Translations []TranslationDetail `json:"translations"`
}

// Clone returns a deep copy.
func (r TranslationsByHashResponse) Clone() TranslationsByHashResponse {
	c := r
	if r.Translations != nil {
		c.Translations = make([]TranslationDetail, len(r.Translations))
		copy(c.Translations, r.Translations)
	}
	return c
}

// TranslationDetail is a single translation entry within a hash response.
type TranslationDetail struct {
	SourceLang     string `json:"source_lang"`
	TargetLang     string `json:"target_lang"`
	TranslatedText string `json:"translated_text"`
	Provider       string `json:"provider"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

// Clone returns a shallow copy.
func (d TranslationDetail) Clone() TranslationDetail {
	return d
}
