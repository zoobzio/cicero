package translator_test

import (
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	pb "github.com/zoobz-io/cicero/proto/translator"
)

// Compile-time assertion: generated server interface is satisfied by the unimplemented base.
var _ pb.TranslatorServiceServer = (*pb.UnimplementedTranslatorServiceServer)(nil)

func TestTranslateRequest_Fields(t *testing.T) {
	req := &pb.TranslateRequest{
		Text:           "Hello, world!",
		SourceLanguage: "en",
		TargetLanguage: "es",
		Route:          "simple",
	}

	if req.GetText() != "Hello, world!" {
		t.Errorf("Text: got %q, want %q", req.GetText(), "Hello, world!")
	}
	if req.GetSourceLanguage() != "en" {
		t.Errorf("SourceLanguage: got %q, want %q", req.GetSourceLanguage(), "en")
	}
	if req.GetTargetLanguage() != "es" {
		t.Errorf("TargetLanguage: got %q, want %q", req.GetTargetLanguage(), "es")
	}
	if req.GetRoute() != "simple" {
		t.Errorf("Route: got %q, want %q", req.GetRoute(), "simple")
	}
}

func TestTranslateResponse_Fields(t *testing.T) {
	resp := &pb.TranslateResponse{
		TranslatedText: "¡Hola, mundo!",
		Provider:       "sidecar",
	}

	if resp.GetTranslatedText() != "¡Hola, mundo!" {
		t.Errorf("TranslatedText: got %q, want %q", resp.GetTranslatedText(), "¡Hola, mundo!")
	}
	if resp.GetProvider() != "sidecar" {
		t.Errorf("Provider: got %q, want %q", resp.GetProvider(), "sidecar")
	}
}

func TestTranslateRequest_ZeroValues(t *testing.T) {
	req := &pb.TranslateRequest{}

	if req.GetText() != "" {
		t.Errorf("zero Text: got %q, want empty", req.GetText())
	}
	if req.GetSourceLanguage() != "" {
		t.Errorf("zero SourceLanguage: got %q, want empty", req.GetSourceLanguage())
	}
	if req.GetTargetLanguage() != "" {
		t.Errorf("zero TargetLanguage: got %q, want empty", req.GetTargetLanguage())
	}
	if req.GetRoute() != "" {
		t.Errorf("zero Route: got %q, want empty", req.GetRoute())
	}
}

func TestTranslateResponse_ZeroValues(t *testing.T) {
	resp := &pb.TranslateResponse{}

	if resp.GetTranslatedText() != "" {
		t.Errorf("zero TranslatedText: got %q, want empty", resp.GetTranslatedText())
	}
	if resp.GetProvider() != "" {
		t.Errorf("zero Provider: got %q, want empty", resp.GetProvider())
	}
}

func TestTranslateRequest_ProtoRoundTrip(t *testing.T) {
	original := &pb.TranslateRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "fr",
		Route:          "simple",
	}

	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}

	decoded := &pb.TranslateRequest{}
	if err := proto.Unmarshal(data, decoded); err != nil {
		t.Fatalf("proto.Unmarshal: %v", err)
	}

	if !proto.Equal(original, decoded) {
		t.Errorf("round-trip mismatch: got %v, want %v", decoded, original)
	}
}

func TestTranslateResponse_ProtoRoundTrip(t *testing.T) {
	original := &pb.TranslateResponse{
		TranslatedText: "Bonjour",
		Provider:       "sidecar",
	}

	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}

	decoded := &pb.TranslateResponse{}
	if err := proto.Unmarshal(data, decoded); err != nil {
		t.Fatalf("proto.Unmarshal: %v", err)
	}

	if !proto.Equal(original, decoded) {
		t.Errorf("round-trip mismatch: got %v, want %v", decoded, original)
	}
}

func TestServiceDescriptor_Present(t *testing.T) {
	desc := pb.TranslatorService_ServiceDesc

	if desc.ServiceName != "translator.TranslatorService" {
		t.Errorf("ServiceName: got %q, want %q", desc.ServiceName, "translator.TranslatorService")
	}
	if len(desc.Methods) != 1 {
		t.Errorf("Methods count: got %d, want 1", len(desc.Methods))
	}
	if desc.Methods[0].MethodName != "Translate" {
		t.Errorf("Method[0].MethodName: got %q, want %q", desc.Methods[0].MethodName, "Translate")
	}
}

func TestServiceDescriptor_ClientInterface(t *testing.T) {
	// Verify the client constructor is callable with a nil connection (compile-time interface check).
	// We pass a nil conn — we only care that the type satisfies the interface, not that it works.
	var conn grpc.ClientConnInterface
	client := pb.NewTranslatorServiceClient(conn)
	if client == nil {
		t.Error("NewTranslatorServiceClient returned nil")
	}
}
