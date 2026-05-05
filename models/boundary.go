package models

import "github.com/zoobz-io/sum"

// RegisterBoundaries registers model boundary processors with the service registry.
// No model boundaries are required in the current implementation.
func RegisterBoundaries(_ sum.Key) error {
	return nil
}
