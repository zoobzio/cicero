package handlers

import "github.com/zoobzio/rocco"

// All returns all public API handlers for registration with the router.
func All() []rocco.Endpoint {
	return []rocco.Endpoint{
		CreateTranslation,
		GetTranslationsByHash,
	}
}
