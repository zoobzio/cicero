//go:build testing

package translator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zoobzio/cicero/api/contracts"
)

// Compile-time assertion: Client satisfies contracts.Translator.
var _ contracts.Translator = (*Client)(nil)

func TestClient_Translate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s, want POST", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/translate") {
			t.Errorf("path: got %s, want /translate", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type: got %s, want application/json", r.Header.Get("Content-Type"))
		}

		var req translateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if req.Q == "" {
			t.Error("request Q is empty")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(translateResponse{TranslatedText: "¡Hola, mundo!"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.Translate(ctx, "Hello, world!", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "¡Hola, mundo!" {
		t.Errorf("result: got %q, want %q", result, "¡Hola, mundo!")
	}
}

func TestClient_Translate_LibreTranslateError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(translateErrorResponse{Error: "Unknown language code"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Translate(ctx, "Hello", "en", "xx")
	if err == nil {
		t.Fatal("expected error for bad language code, got nil")
	}
	if !strings.Contains(err.Error(), "Unknown language code") {
		t.Errorf("error message: got %q, want to contain %q", err.Error(), "Unknown language code")
	}
}

func TestClient_Translate_UnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Translate(ctx, "Hello", "en", "es")
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
}

func TestClient_Translate_RequestBody(t *testing.T) {
	var captured translateRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(translateResponse{TranslatedText: "Bonjour"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Translate(ctx, "Hello", "en", "fr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured.Q != "Hello" {
		t.Errorf("request Q: got %q, want %q", captured.Q, "Hello")
	}
	if captured.Source != "en" {
		t.Errorf("request Source: got %q, want %q", captured.Source, "en")
	}
	if captured.Target != "fr" {
		t.Errorf("request Target: got %q, want %q", captured.Target, "fr")
	}
}

func TestClient_Close(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(translateResponse{TranslatedText: "test"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	if err := client.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}
