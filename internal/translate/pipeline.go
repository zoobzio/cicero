package translate

import "github.com/zoobzio/pipz"

// Pipeline identities for observability and debugging.
var (
	PipelineID    = pipz.NewIdentity("translate.pipeline", "Translation pipeline")
	HashID        = pipz.NewIdentity("translate.hash", "Compute content hash")
	DeduplicateID = pipz.NewIdentity("translate.dedup", "Check for existing translation")
	ClassifyID    = pipz.NewIdentity("translate.classify", "Classify content complexity")
	TranslateID   = pipz.NewIdentity("translate.translate", "Translate via provider")
	StoreID       = pipz.NewIdentity("translate.store", "Persist source and translation")
)

// NewPipeline constructs the translation pipeline.
// Stages execute sequentially: hash -> deduplicate -> classify -> translate -> store.
// If deduplication finds an existing translation, the translate and store stages
// are skipped (they check job.Existing and return early).
func NewPipeline() *pipz.Sequence[*Job] {
	return pipz.NewSequence(PipelineID,
		pipz.Apply(HashID, hashStage),
		pipz.Apply(DeduplicateID, deduplicateStage),
		pipz.Apply(ClassifyID, classifyStage),
		pipz.Apply(TranslateID, translateStage),
		pipz.Apply(StoreID, storeStage),
	)
}
