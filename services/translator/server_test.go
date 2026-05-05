package translator_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	pb "github.com/zoobz-io/cicero/proto/translator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	. "github.com/zoobz-io/cicero/translator"
)

const bufSize = 1024 * 1024

// setupGRPC starts an in-process gRPC server with the given Server and returns
// a client connected to it via bufconn. Caller must defer the returned cleanup.
func setupGRPC(t *testing.T, srv *Server) (pb.TranslatorServiceClient, func()) {
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

	cleanup := func() {
		_ = conn.Close()
		s.Stop()
		_ = lis.Close()
	}

	return pb.NewTranslatorServiceClient(conn), cleanup
}

// newLibreTranslateMock returns a test HTTP server simulating LibreTranslate.
func newLibreTranslateMock(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// --- Proxy logic tests (Server.Translate / callLibreTranslate) ---

func TestServer_Translate_SimpleRoute_Success(t *testing.T) {
	mock := newLibreTranslateMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"translatedText": "¡Hola, mundo!"})
	})

	srv := NewServer(mock.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := srv.Translate(ctx, &pb.TranslateRequest{
		Text:           "Hello, world!",
		SourceLanguage: "en",
		TargetLanguage: "es",
		Route:          "simple",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TranslatedText != "¡Hola, mundo!" {
		t.Errorf("TranslatedText: got %q, want %q", resp.TranslatedText, "¡Hola, mundo!")
	}
	if resp.Provider != "libretranslate" {
		t.Errorf("Provider: got %q, want %q", resp.Provider, "libretranslate")
	}
}

func TestServer_Translate_LibreTranslateRequestFields(t *testing.T) {
	var body map[string]string

	mock := newLibreTranslateMock(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"translatedText": "Bonjour"})
	})

	srv := NewServer(mock.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := srv.Translate(ctx, &pb.TranslateRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "fr",
		Route:          "simple",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body["q"] != "Hello" {
		t.Errorf("q: got %q, want %q", body["q"], "Hello")
	}
	if body["source"] != "en" {
		t.Errorf("source: got %q, want %q", body["source"], "en")
	}
	if body["target"] != "fr" {
		t.Errorf("target: got %q, want %q", body["target"], "fr")
	}
}

func TestServer_Translate_ComplexRoute_Unimplemented(t *testing.T) {
	srv := NewServer("http://unused")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := srv.Translate(ctx, &pb.TranslateRequest{
		Text:           "Nuanced text",
		SourceLanguage: "en",
		TargetLanguage: "es",
		Route:          "complex",
	})
	if err == nil {
		t.Fatal("expected error for complex route, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T: %v", err, err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("code: got %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestServer_Translate_LibreTranslateBadRequest(t *testing.T) {
	mock := newLibreTranslateMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Unknown language code"})
	})

	srv := NewServer(mock.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := srv.Translate(ctx, &pb.TranslateRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "xx",
		Route:          "simple",
	})
	if err == nil {
		t.Fatal("expected error for bad language code, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T: %v", err, err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code: got %v, want %v", st.Code(), codes.InvalidArgument)
	}
	if !strings.Contains(st.Message(), "Unknown language code") {
		t.Errorf("message: got %q, want to contain %q", st.Message(), "Unknown language code")
	}
}

func TestServer_Translate_LibreTranslateUnavailable(t *testing.T) {
	// Point at a port that refuses connections.
	srv := NewServer("http://127.0.0.1:1")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := srv.Translate(ctx, &pb.TranslateRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "es",
		Route:          "simple",
	})
	if err == nil {
		t.Fatal("expected error when libretranslate is unavailable, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T: %v", err, err)
	}
	if st.Code() != codes.Unavailable {
		t.Errorf("code: got %v, want %v", st.Code(), codes.Unavailable)
	}
}

func TestServer_Translate_LibreTranslateUnexpectedStatus(t *testing.T) {
	mock := newLibreTranslateMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	srv := NewServer(mock.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := srv.Translate(ctx, &pb.TranslateRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "es",
		Route:          "simple",
	})
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T: %v", err, err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("code: got %v, want %v", st.Code(), codes.Internal)
	}
}

// --- gRPC wiring tests (end-to-end via bufconn) ---

func TestGRPCServer_Translate_Success(t *testing.T) {
	mock := newLibreTranslateMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"translatedText": "¡Hola!"})
	})

	grpcClient, cleanup := setupGRPC(t, NewServer(mock.URL))
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := grpcClient.Translate(ctx, &pb.TranslateRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "es",
		Route:          "simple",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TranslatedText != "¡Hola!" {
		t.Errorf("TranslatedText: got %q, want %q", resp.TranslatedText, "¡Hola!")
	}
	if resp.Provider != "libretranslate" {
		t.Errorf("Provider: got %q, want %q", resp.Provider, "libretranslate")
	}
}

func TestGRPCServer_Translate_ComplexRoute_Unimplemented(t *testing.T) {
	grpcClient, cleanup := setupGRPC(t, NewServer("http://unused"))
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := grpcClient.Translate(ctx, &pb.TranslateRequest{
		Text:           "Complex",
		SourceLanguage: "en",
		TargetLanguage: "es",
		Route:          "complex",
	})
	if err == nil {
		t.Fatal("expected error for complex route, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T: %v", err, err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("code: got %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestGRPCServer_Translate_LibreTranslateError_PropagatesStatus(t *testing.T) {
	mock := newLibreTranslateMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Unknown language code"})
	})

	grpcClient, cleanup := setupGRPC(t, NewServer(mock.URL))
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := grpcClient.Translate(ctx, &pb.TranslateRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "xx",
		Route:          "simple",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T: %v", err, err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code: got %v, want %v", st.Code(), codes.InvalidArgument)
	}
}
