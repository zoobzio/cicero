// Package translate provides the translation pipeline for cicero.
// The pipeline processes translation requests through five stages:
// hash, deduplicate, classify, translate, and store.
package translate

import "github.com/zoobzio/cicero/models"

// Job is the pipeline carrier. It flows through every stage, accumulating
// state as processing progresses.
type Job struct {
	// Dedup — set by the deduplicate stage if a cached translation exists.
	Existing *models.Translation

	// Input — set by the caller before pipeline execution.
	SourceText string
	SourceLang string
	TargetLang string
	TenantID   string

	// Computed — populated by pipeline stages.
	Hash           string
	TranslatedText string
	Provider       string
	Status         string
	Classification models.Classification
}

// Clone returns a deep copy of the job for parallel processing.
func (j *Job) Clone() *Job {
	c := *j
	if j.Existing != nil {
		e := *j.Existing
		c.Existing = &e
	}
	if j.Classification.Signals != nil {
		c.Classification.Signals = make([]string, len(j.Classification.Signals))
		copy(c.Classification.Signals, j.Classification.Signals)
	}
	return &c
}
