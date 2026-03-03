package wire

import "github.com/zoobzio/sum"

// RegisterBoundaries registers wire boundary processors with the service registry.
// No wire boundaries are required in the current implementation — public API wire
// types do not contain sensitive fields requiring masking or hashing.
func RegisterBoundaries(_ sum.Key) error {
	return nil
}
