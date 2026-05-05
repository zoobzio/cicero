//go:build testing

package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/cicero/internal/translate"
	"github.com/zoobz-io/cicero/models"
	cicerotest "github.com/zoobz-io/cicero/testing"
	"github.com/zoobz-io/pipz"
	roccotest "github.com/zoobz-io/rocco/testing"
	"github.com/zoobz-io/sum"
)

func TestCreateTranslation_Success(t *testing.T) {
	pipeline := &cicerotest.MockPipeline{
		OnProcess: func(_ context.Context, job *translate.Job) (*translate.Job, error) {
			job.Hash = "315f5bdb76d078c43b8ac0064e4a0164"
			job.TranslatedText = "¡Hola, mundo!"
			job.Provider = "sidecar"
			job.Status = "completed"
			job.Classification = models.Classification{Route: models.RouteSimple}
			return job, nil
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[pipz.Chainable[*translate.Job]](k, pipeline)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)

	engine := roccotest.TestEngine()
	engine.WithHandlers(CreateTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations", map[string]string{
		"text":        "Hello, world!",
		"source_lang": "en",
		"target_lang": "es",
		"tenant_id":   "zoobzio",
	})

	if resp.StatusCode() != http.StatusCreated {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusCreated, resp.BodyString())
	}

	var body map[string]any
	if err := resp.DecodeJSON(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	checks := []struct {
		field string
		want  string
	}{
		{"hash", "315f5bdb76d078c43b8ac0064e4a0164"},
		{"translated_text", "¡Hola, mundo!"},
		{"provider", "sidecar"},
		{"status", "completed"},
		{"classification", "simple"},
		{"source_lang", "en"},
		{"target_lang", "es"},
	}
	for _, c := range checks {
		got, _ := body[c.field].(string)
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.field, got, c.want)
		}
	}
}

func TestCreateTranslation_PipelineError_Returns500(t *testing.T) {
	pipeline := &cicerotest.MockPipeline{
		OnProcess: func(_ context.Context, job *translate.Job) (*translate.Job, error) {
			return nil, errors.New("sidecar unavailable")
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[pipz.Chainable[*translate.Job]](k, pipeline)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)

	engine := roccotest.TestEngine()
	engine.WithHandlers(CreateTranslation)

	resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations", map[string]string{
		"text":        "Hello",
		"source_lang": "en",
		"target_lang": "es",
		"tenant_id":   "zoobzio",
	})

	if resp.StatusCode() != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusInternalServerError, resp.BodyString())
	}
}

func TestCreateTranslation_ValidationError_MissingText(t *testing.T) {
	pipeline := &cicerotest.MockPipeline{}

	sum.Reset()
	k := sum.Start()
	sum.Register[pipz.Chainable[*translate.Job]](k, pipeline)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)

	engine := roccotest.TestEngine()
	engine.WithHandlers(CreateTranslation)

	tests := []struct {
		name string
		body map[string]string
	}{
		{name: "missing text", body: map[string]string{"source_lang": "en", "target_lang": "es", "tenant_id": "zoobzio"}},
		{name: "missing source_lang", body: map[string]string{"text": "Hello", "target_lang": "es", "tenant_id": "zoobzio"}},
		{name: "missing target_lang", body: map[string]string{"text": "Hello", "source_lang": "en", "tenant_id": "zoobzio"}},
		{name: "missing tenant_id", body: map[string]string{"text": "Hello", "source_lang": "en", "target_lang": "es"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := roccotest.ServeRequest(engine, http.MethodPost, "/translations", tc.body)
			if resp.StatusCode() == http.StatusCreated {
				t.Errorf("expected non-201 for %s, got 201", tc.name)
			}
		})
	}
}

func TestGetTranslationsByHash_Success(t *testing.T) {
	src := &models.Source{
		Hash:     "315f5bdb76d078c43b8ac0064e4a0164",
		Text:     "Hello, world!",
		TenantID: "zoobzio",
	}
	translationList := []*models.Translation{
		{
			SourceLang: "en",
			TargetLang: "es",
			Text:       "¡Hola, mundo!",
			Provider:   "sidecar",
			Status:     "completed",
			CreatedAt:  time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		},
	}

	ms := &cicerotest.MockSources{
		OnGet: func(_ context.Context, _ string) (*models.Source, error) {
			return src, nil
		},
	}
	mt := &cicerotest.MockTranslations{
		OnListBySourceHash: func(_ context.Context, _ string) ([]*models.Translation, error) {
			return translationList, nil
		},
	}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Sources](k, ms)
	sum.Register[contracts.Translations](k, mt)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)

	engine := roccotest.TestEngine()
	engine.WithHandlers(GetTranslationsByHash)

	resp := roccotest.ServeRequest(engine, http.MethodGet, "/translations/315f5bdb76d078c43b8ac0064e4a0164", nil)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusOK, resp.BodyString())
	}

	var body map[string]any
	if err := resp.DecodeJSON(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["hash"] != src.Hash {
		t.Errorf("hash: got %v, want %q", body["hash"], src.Hash)
	}
	if body["source_text"] != src.Text {
		t.Errorf("source_text: got %v, want %q", body["source_text"], src.Text)
	}

	txs, _ := body["translations"].([]any)
	if len(txs) != 1 {
		t.Errorf("translations length: got %d, want 1", len(txs))
	}
}

func TestGetTranslationsByHash_SourceNotFound_Returns404(t *testing.T) {
	ms := &cicerotest.MockSources{
		OnGet: func(_ context.Context, _ string) (*models.Source, error) {
			return nil, errors.New("not found")
		},
	}
	mt := &cicerotest.MockTranslations{}

	sum.Reset()
	k := sum.Start()
	sum.Register[contracts.Sources](k, ms)
	sum.Register[contracts.Translations](k, mt)
	sum.Freeze(k)
	t.Cleanup(sum.Reset)

	engine := roccotest.TestEngine()
	engine.WithHandlers(GetTranslationsByHash)

	resp := roccotest.ServeRequest(engine, http.MethodGet, "/translations/doesnotexist", nil)

	if resp.StatusCode() != http.StatusNotFound {
		t.Errorf("status: got %d, want %d (body: %s)", resp.StatusCode(), http.StatusNotFound, resp.BodyString())
	}
}
