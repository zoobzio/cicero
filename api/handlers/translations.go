package handlers

import (
	"github.com/zoobzio/rocco"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/cicero/api/contracts"
	"github.com/zoobzio/cicero/api/transformers"
	"github.com/zoobzio/cicero/api/wire"
	"github.com/zoobzio/cicero/internal/translate"
	"github.com/zoobzio/cicero/models"
	"github.com/zoobzio/pipz"
)

// CreateTranslation submits text for translation.
// Hashes the source text, deduplicates, classifies, translates via the sidecar,
// stores the result, and returns the content hash and translation.
var CreateTranslation = rocco.POST("/translations", func(req *rocco.Request[wire.TranslateRequest]) (wire.TranslateResponse, error) {
	pipeline := sum.MustUse[pipz.Chainable[*translate.Job]](req.Context)

	job := &translate.Job{
		SourceText: req.Body.Text,
		SourceLang: req.Body.SourceLang,
		TargetLang: req.Body.TargetLang,
		TenantID:   req.Body.TenantID,
	}

	result, err := pipeline.Process(req.Context, job)
	if err != nil {
		return wire.TranslateResponse{}, ErrTranslationFailed.WithCause(err)
	}

	src := &models.Source{
		Hash:     result.Hash,
		Text:     result.SourceText,
		TenantID: result.TenantID,
	}

	translation := &models.Translation{
		SourceHash: result.Hash,
		SourceLang: result.SourceLang,
		TargetLang: result.TargetLang,
		Text:       result.TranslatedText,
		Provider:   result.Provider,
		Status:     result.Status,
		TenantID:   result.TenantID,
	}

	return transformers.SourceAndTranslationToResponse(src, translation, result.Classification.Route), nil
}).WithSummary("Submit translation").
	WithDescription("Submits text for translation. Returns the content hash and translation result.").
	WithTags("Translations").
	WithErrors(ErrTranslationFailed).
	WithSuccessStatus(201)

// GetTranslationsByHash retrieves the source text and all translations for a given content hash.
var GetTranslationsByHash = rocco.GET("/translations/{hash}", func(req *rocco.Request[rocco.NoBody]) (wire.TranslationsByHashResponse, error) {
	hash := req.Params.Path["hash"]

	sources := sum.MustUse[contracts.Sources](req.Context)
	src, err := sources.Get(req.Context, hash)
	if err != nil {
		return wire.TranslationsByHashResponse{}, ErrSourceNotFound
	}

	translations := sum.MustUse[contracts.Translations](req.Context)
	list, err := translations.ListBySourceHash(req.Context, hash)
	if err != nil {
		return wire.TranslationsByHashResponse{}, err
	}

	return transformers.SourceAndTranslationsToHashResponse(src, list), nil
}).WithPathParams("hash").
	WithSummary("Get translations by hash").
	WithDescription("Retrieves the source text and all translations for a given content hash.").
	WithTags("Translations").
	WithErrors(ErrSourceNotFound)
