package translate

import (
	"context"

	"github.com/zoobzio/cicero/internal/classify"
	"github.com/zoobzio/sum"
)

// classifyStage runs the complexity classifier against the source text.
// The classification result determines how the text will be routed.
// In the first slice, the classifier always returns Simple (route to sidecar).
func classifyStage(ctx context.Context, job *Job) (*Job, error) {
	classifier := sum.MustUse[classify.Classifier](ctx)

	result, err := classifier.Classify(ctx, job.SourceText)
	if err != nil {
		return job, err
	}

	job.Classification = result
	return job, nil
}
