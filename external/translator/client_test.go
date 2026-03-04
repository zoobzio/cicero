//go:build testing

package translator

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/zoobzio/cicero/api/contracts"
	"github.com/zoobzio/cicero/models"
	pb "github.com/zoobzio/cicero/proto/translator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// Compile-time assertion: Client satisfies contracts.Translator.
var _ contracts.Translator = (*Client)(nil)

const bufSize = 1024 * 1024

// fakeServer is a test implementation of TranslatorServiceServer.
type fakeServer struct {
	pb.UnimplementedTranslatorServiceServer
	translate func(ctx context.Context, req *pb.TranslateRequest) (*pb.TranslateResponse, error)
}

func (f *fakeServer) Translate(ctx context.Context, req *pb.TranslateRequest) (*pb.TranslateResponse, error) {
	if f.translate != nil {
		return f.translate(ctx, req)
	}
	return &pb.TranslateResponse{
		TranslatedText: "default",
		Provider:       "sidecar",
	}, nil
}

// setupBufconn starts an in-process gRPC server using bufconn and returns a
// Client wired to it. The caller must defer the returned cleanup func.
func setupBufconn(t *testing.T, srv pb.TranslatorServiceServer) (*Client, func()) {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterTranslatorServiceServer(s, srv)

	go func() { _ = s.Serve(lis) }()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("bufconn dial: %v", err)
	}

	c := &Client{
		addr: "bufnet",
		conn: conn,
	}
	c.pipeline = c.buildPipeline()

	cleanup := func() {
		_ = c.Close()
		s.Stop()
		_ = lis.Close()
	}

	return c, cleanup
}

func TestClient_Translate_Success(t *testing.T) {
	srv := &fakeServer{
		translate: func(_ context.Context, _ *pb.TranslateRequest) (*pb.TranslateResponse, error) {
			return &pb.TranslateResponse{
				TranslatedText: "¡Hola, mundo!",
				Provider:       "sidecar",
			}, nil
		},
	}

	client, cleanup := setupBufconn(t, srv)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, provider, err := client.Translate(ctx, "Hello, world!", "en", "es", models.RouteSimple)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "¡Hola, mundo!" {
		t.Errorf("result: got %q, want %q", result, "¡Hola, mundo!")
	}
	if provider != "sidecar" {
		t.Errorf("provider: got %q, want %q", provider, "sidecar")
	}
}

func TestClient_Translate_RequestFields(t *testing.T) {
	var captured *pb.TranslateRequest

	srv := &fakeServer{
		translate: func(_ context.Context, req *pb.TranslateRequest) (*pb.TranslateResponse, error) {
			captured = req
			return &pb.TranslateResponse{TranslatedText: "Bonjour", Provider: "sidecar"}, nil
		},
	}

	client, cleanup := setupBufconn(t, srv)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := client.Translate(ctx, "Hello", "en", "fr", models.RouteSimple)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured == nil {
		t.Fatal("server received no request")
	}
	if captured.Text != "Hello" {
		t.Errorf("Text: got %q, want %q", captured.Text, "Hello")
	}
	if captured.SourceLanguage != "en" {
		t.Errorf("SourceLanguage: got %q, want %q", captured.SourceLanguage, "en")
	}
	if captured.TargetLanguage != "fr" {
		t.Errorf("TargetLanguage: got %q, want %q", captured.TargetLanguage, "fr")
	}
	if captured.Route != string(models.RouteSimple) {
		t.Errorf("Route: got %q, want %q", captured.Route, models.RouteSimple)
	}
}

func TestClient_Translate_RouteForwarded(t *testing.T) {
	var capturedRoute string

	srv := &fakeServer{
		translate: func(_ context.Context, req *pb.TranslateRequest) (*pb.TranslateResponse, error) {
			capturedRoute = req.Route
			return &pb.TranslateResponse{TranslatedText: "Hola", Provider: "llm"}, nil
		},
	}

	client, cleanup := setupBufconn(t, srv)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, provider, err := client.Translate(ctx, "Complex text", "en", "es", models.RouteComplex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedRoute != string(models.RouteComplex) {
		t.Errorf("Route forwarded: got %q, want %q", capturedRoute, models.RouteComplex)
	}
	if provider != "llm" {
		t.Errorf("provider: got %q, want %q", provider, "llm")
	}
}

func TestClient_Translate_ServerError(t *testing.T) {
	srv := &fakeServer{
		translate: func(_ context.Context, _ *pb.TranslateRequest) (*pb.TranslateResponse, error) {
			return nil, status.Error(codes.Internal, "upstream unavailable")
		},
	}

	client, cleanup := setupBufconn(t, srv)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := client.Translate(ctx, "Hello", "en", "es", models.RouteSimple)
	if err == nil {
		t.Fatal("expected error from server, got nil")
	}
}

func TestClient_Close(t *testing.T) {
	srv := &fakeServer{}

	client, cleanup := setupBufconn(t, srv)
	_ = cleanup // Close called explicitly below

	if err := client.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}
