package translate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// hashStage computes the SHA-256 content hash of the source text,
// truncated to 32 hex characters (16 bytes). This hash serves as
// the content-addressable key for deduplication and retrieval.
func hashStage(_ context.Context, job *Job) (*Job, error) {
	sum := sha256.Sum256([]byte(job.SourceText))
	job.Hash = hex.EncodeToString(sum[:])[:32]
	return job, nil
}
