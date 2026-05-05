package events

import (
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/sum"
)

// Translation signals for domain lifecycle.
var (
	translationCompletedSignal = capitan.NewSignal("cicero.translation.completed", "Translation completed successfully")
	translationFailedSignal    = capitan.NewSignal("cicero.translation.failed", "Translation failed")
	translationCachedSignal    = capitan.NewSignal("cicero.translation.cached", "Existing translation served from cache")
)

// TranslationEvent carries contextual data for translation lifecycle signals.
type TranslationEvent struct {
	Hash       string `json:"hash"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	Provider   string `json:"provider"`
	Status     string `json:"status"`
	Cached     bool   `json:"cached"`
}

// Translation groups the domain lifecycle events for translations.
var Translation = struct {
	Completed sum.Event[TranslationEvent]
	Failed    sum.Event[TranslationEvent]
	Cached    sum.Event[TranslationEvent]
}{
	Completed: sum.NewInfoEvent[TranslationEvent](translationCompletedSignal),
	Failed:    sum.NewWarnEvent[TranslationEvent](translationFailedSignal),
	Cached:    sum.NewInfoEvent[TranslationEvent](translationCachedSignal),
}
