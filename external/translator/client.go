// Package translator provides a gRPC client for the translator sidecar.
package translator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zoobz-io/cicero/models"
	pb "github.com/zoobz-io/cicero/proto/translator"
	"github.com/zoobz-io/pipz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Resilience configuration.
const (
	translateTimeout          = 30 * time.Second
	translateMaxAttempts      = 3
	translateBackoffDelay     = 200 * time.Millisecond
	translateFailureThreshold = 5
	translateResetTimeout     = 30 * time.Second
)

// Pipeline identities for the resilience stack.
var (
	translateProcessorID = pipz.NewIdentity("translator.call", "gRPC call to translator sidecar")
	translateTimeoutID   = pipz.NewIdentity("translator.timeout", "Timeout for translation calls")
	translateBackoffID   = pipz.NewIdentity("translator.backoff", "Backoff retry for translation calls")
	translateBreakerID   = pipz.NewIdentity("translator.breaker", "Circuit breaker for translator sidecar")
)

// translateCall is the pipeline carrier for a single translation request.
type translateCall struct {
	text       string
	sourceLang string
	targetLang string
	route      models.Route
	result     string
	provider   string
}

func (c *translateCall) Clone() *translateCall {
	clone := *c
	return &clone
}

// Client calls the translator gRPC sidecar with a resilient pipeline.
// The pipeline stacks CircuitBreaker -> Backoff -> Timeout -> gRPC Processor.
// Create one Client per service and reuse it — the circuit breaker is stateful.
type Client struct {
	pipeline pipz.Chainable[*translateCall]
	conn     *grpc.ClientConn
	addr     string
	mu       sync.Mutex
}

// NewClient creates a translator gRPC client targeting the given address.
// The addr should be a host:port string (e.g., "localhost:9091").
func NewClient(addr string) *Client {
	c := &Client{addr: addr}
	c.pipeline = c.buildPipeline()
	return c
}

// buildPipeline constructs the resilient gRPC processing pipeline.
func (c *Client) buildPipeline() pipz.Chainable[*translateCall] {
	processor := pipz.Apply(translateProcessorID, c.doTranslate)

	return pipz.NewCircuitBreaker(translateBreakerID,
		pipz.NewBackoff(translateBackoffID,
			pipz.NewTimeout(translateTimeoutID, processor, translateTimeout),
			translateMaxAttempts, translateBackoffDelay,
		),
		translateFailureThreshold, translateResetTimeout,
	)
}

// doTranslate performs the actual gRPC call to the translator sidecar.
func (c *Client) doTranslate(ctx context.Context, call *translateCall) (*translateCall, error) {
	client, err := c.dial()
	if err != nil {
		return call, err
	}

	res, err := client.Translate(ctx, &pb.TranslateRequest{
		Text:           call.text,
		SourceLanguage: call.sourceLang,
		TargetLanguage: call.targetLang,
		Route:          string(call.route),
	})
	if err != nil {
		return call, fmt.Errorf("translator: %w", err)
	}

	call.result = res.TranslatedText
	call.provider = res.Provider
	return call, nil
}

// dial lazily establishes a gRPC connection to the translator sidecar.
func (c *Client) dial() (pb.TranslatorServiceClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return pb.NewTranslatorServiceClient(c.conn), nil
	}

	conn, err := grpc.NewClient(c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial translator at %s: %w", c.addr, err)
	}

	c.conn = conn
	return pb.NewTranslatorServiceClient(conn), nil
}

// Translate sends text to the translator sidecar and returns the translated string
// and the provider name that handled the request.
func (c *Client) Translate(ctx context.Context, text, sourceLang, targetLang string, route models.Route) (result string, provider string, err error) {
	call := &translateCall{
		text:       text,
		sourceLang: sourceLang,
		targetLang: targetLang,
		route:      route,
	}
	out, err := c.pipeline.Process(ctx, call)
	if err != nil {
		return "", "", err
	}
	return out.result, out.provider, nil
}

// Close shuts down the gRPC connection and resilience pipeline.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	if c.pipeline != nil {
		if err := c.pipeline.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, err)
		}
		c.conn = nil
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
