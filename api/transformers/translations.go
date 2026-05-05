// Package transformers provides pure functions for mapping between domain models and API wire types.
package transformers

import (
	"github.com/zoobz-io/cicero/api/wire"
	"github.com/zoobz-io/cicero/models"
)

// SourceAndTranslationToResponse maps a source and translation model to a TranslateResponse.
// The classification route is passed separately as it comes from the pipeline job.
func SourceAndTranslationToResponse(source *models.Source, translation *models.Translation, classification models.Route) wire.TranslateResponse {
	return wire.TranslateResponse{
		Hash:           source.Hash,
		SourceText:     source.Text,
		TranslatedText: translation.Text,
		SourceLang:     translation.SourceLang,
		TargetLang:     translation.TargetLang,
		Classification: string(classification),
		Provider:       translation.Provider,
		Status:         translation.Status,
	}
}

// SourceAndTranslationsToHashResponse maps a source and its translations to a TranslationsByHashResponse.
func SourceAndTranslationsToHashResponse(source *models.Source, translations []*models.Translation) wire.TranslationsByHashResponse {
	details := make([]wire.TranslationDetail, len(translations))
	for i, t := range translations {
		details[i] = TranslationToDetail(t)
	}
	return wire.TranslationsByHashResponse{
		Hash:         source.Hash,
		SourceText:   source.Text,
		Translations: details,
	}
}

// TranslationToDetail maps a single translation model to a TranslationDetail.
func TranslationToDetail(t *models.Translation) wire.TranslationDetail {
	return wire.TranslationDetail{
		SourceLang:     t.SourceLang,
		TargetLang:     t.TargetLang,
		TranslatedText: t.Text,
		Provider:       t.Provider,
		Status:         t.Status,
		CreatedAt:      t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
