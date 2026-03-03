//go:build testing

package translate

import (
	"context"
	"testing"
)

func TestHashStage_KnownVector(t *testing.T) {
	// SHA-256("Hello, world!") truncated to 32 hex chars.
	// Verified via Go: sha256.Sum256([]byte("Hello, world!"))[:16] hex-encoded.
	const want = "315f5bdb76d078c43b8ac0064e4a0164"

	job := &Job{SourceText: "Hello, world!"}
	result, err := hashStage(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Hash != want {
		t.Errorf("hash: got %q, want %q", result.Hash, want)
	}
}

func TestHashStage_HashLength(t *testing.T) {
	job := &Job{SourceText: "any text"}
	result, err := hashStage(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hash) != 32 {
		t.Errorf("hash length: got %d, want 32", len(result.Hash))
	}
}

func TestHashStage_Deterministic(t *testing.T) {
	job1 := &Job{SourceText: "same input"}
	job2 := &Job{SourceText: "same input"}

	r1, _ := hashStage(context.Background(), job1)
	r2, _ := hashStage(context.Background(), job2)

	if r1.Hash != r2.Hash {
		t.Errorf("same input produced different hashes: %q vs %q", r1.Hash, r2.Hash)
	}
}

func TestHashStage_DifferentInputs(t *testing.T) {
	job1 := &Job{SourceText: "input one"}
	job2 := &Job{SourceText: "input two"}

	r1, _ := hashStage(context.Background(), job1)
	r2, _ := hashStage(context.Background(), job2)

	if r1.Hash == r2.Hash {
		t.Error("different inputs produced identical hash")
	}
}

func TestHashStage_EmptyString(t *testing.T) {
	job := &Job{SourceText: ""}
	result, err := hashStage(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error on empty string: %v", err)
	}
	if len(result.Hash) != 32 {
		t.Errorf("hash length for empty string: got %d, want 32", len(result.Hash))
	}
}
