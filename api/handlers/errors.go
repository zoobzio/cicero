// Package handlers provides HTTP handlers for the public API.
package handlers

import "github.com/zoobzio/rocco"

// Domain errors for the public API.
var (
	// ErrSourceNotFound is returned when a source hash has no matching record.
	ErrSourceNotFound = rocco.ErrNotFound.WithMessage("source not found")
	// ErrTranslationFailed is returned when the translation provider returns an error.
	ErrTranslationFailed = rocco.ErrInternalServer.WithMessage("translation failed")
)
