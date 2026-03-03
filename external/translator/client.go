// Package translator provides an HTTP client for the LibreTranslate sidecar.
package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zoobzio/pipz"
)

// Pipeline identities for the resilience stack.
var (
	translateProcessorID = pipz.NewIdentity("translator.call", "HTTP call to LibreTranslate")
	translateTimeoutID   = pipz.NewIdentity("translator.timeout", "Timeout for translation calls")
	translateBackoffID   = pipz.NewIdentity("translator.backoff", "Backoff retry for translation calls")
	translateBreakerID   = pipz.NewIdentity("translator.breaker", "Circuit breaker for LibreTranslate")
)

// translateCall is the pipeline carrier for a single translation request.
type translateCall struct {
	text       string
	sourceLang string
	targetLang string
	result     string
}

// Clone returns a copy of the call carrier.
func (c *translateCall) Clone() *translateCall {
	clone := *c
	return &clone
}

// translateRequest is the JSON body sent to LibreTranslate.
type translateRequest struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// translateResponse is the JSON body returned by LibreTranslate on success.
type translateResponse struct {
	TranslatedText string `json:"translatedText"`
}

// translateErrorResponse is the JSON body returned by LibreTranslate on error.
type translateErrorResponse struct {
	Error string `json:"error"`
}

// Client calls the LibreTranslate REST API with a resilient pipeline.
// The pipeline stacks CircuitBreaker -> Backoff -> Timeout -> HTTP Processor.
// Create one Client per service and reuse it — the circuit breaker is stateful.
type Client struct {
	pipeline pipz.Chainable[*translateCall]
	httpAddr string
}

// NewClient creates a LibreTranslate client targeting the given address.
// The addr should include scheme and host (e.g., "http://localhost:5000").
func NewClient(addr string) *Client {
	c := &Client{httpAddr: addr}

	httpClient := &http.Client{}

	processor := pipz.Apply(translateProcessorID, func(ctx context.Context, call *translateCall) (*translateCall, error) {
		body, err := json.Marshal(translateRequest{
			Q:      call.text,
			Source: call.sourceLang,
			Target: call.targetLang,
		})
		if err != nil {
			return call, fmt.Errorf("marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr+"/translate", bytes.NewReader(body))
		if err != nil {
			return call, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return call, fmt.Errorf("http post: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errResp translateErrorResponse
			if jsonErr := json.NewDecoder(resp.Body).Decode(&errResp); jsonErr == nil && errResp.Error != "" {
				return call, fmt.Errorf("libretranslate error: %s", errResp.Error)
			}
			return call, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}

		var result translateResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return call, fmt.Errorf("decode response: %w", err)
		}

		call.result = result.TranslatedText
		return call, nil
	})

	withTimeout := pipz.NewTimeout(translateTimeoutID, processor, 30*time.Second)
	withBackoff := pipz.NewBackoff(translateBackoffID, withTimeout, 3, 200*time.Millisecond)
	withBreaker := pipz.NewCircuitBreaker(translateBreakerID, withBackoff, 5, 30*time.Second)

	c.pipeline = withBreaker
	return c
}

// Translate sends text to LibreTranslate and returns the translated string.
func (c *Client) Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	call := &translateCall{
		text:       text,
		sourceLang: sourceLang,
		targetLang: targetLang,
	}
	result, err := c.pipeline.Process(ctx, call)
	if err != nil {
		return "", err
	}
	return result.result, nil
}

// Close shuts down the resilience pipeline.
func (c *Client) Close() error {
	return c.pipeline.Close()
}
