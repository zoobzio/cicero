// Package translator implements the gRPC TranslatorService.
package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	pb "github.com/zoobzio/cicero/proto/translator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const libreTranslateProvider = "libretranslate"

// libreTranslateRequest is the JSON body sent to LibreTranslate.
type libreTranslateRequest struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// libreTranslateResponse is the JSON body returned by LibreTranslate on success.
type libreTranslateResponse struct {
	TranslatedText string `json:"translatedText"`
}

// libreTranslateError is the JSON body returned by LibreTranslate on error.
type libreTranslateError struct {
	Error string `json:"error"`
}

// Server implements TranslatorService. It routes simple requests to LibreTranslate
// over HTTP and returns provider metadata in the response.
type Server struct {
	pb.UnimplementedTranslatorServiceServer
	httpClient    *http.Client
	libreTranslateAddr string
}

// NewServer creates a new TranslatorService server targeting the given LibreTranslate address.
// The addr should include scheme and host (e.g., "http://localhost:5000").
func NewServer(libreTranslateAddr string) *Server {
	return &Server{
		httpClient:         &http.Client{},
		libreTranslateAddr: libreTranslateAddr,
	}
}

// Translate handles a translation request. Simple-routed requests are forwarded to
// LibreTranslate. Complex-routed requests return Unimplemented (LLM provider is a future concern).
func (s *Server) Translate(ctx context.Context, req *pb.TranslateRequest) (*pb.TranslateResponse, error) {
	log.Printf("translate request: route=%s source=%s target=%s", req.Route, req.SourceLanguage, req.TargetLanguage)

	switch req.Route {
	case "complex":
		return nil, status.Errorf(codes.Unimplemented, "complex route not yet supported")
	default:
		// "simple" and unrecognised routes go to LibreTranslate.
	}

	translated, err := s.callLibreTranslate(ctx, req.Text, req.SourceLanguage, req.TargetLanguage)
	if err != nil {
		return nil, err
	}

	log.Printf("translate completed: source=%s target=%s", req.SourceLanguage, req.TargetLanguage)
	return &pb.TranslateResponse{
		TranslatedText: translated,
		Provider:       libreTranslateProvider,
	}, nil
}

// callLibreTranslate sends a translation request to the LibreTranslate REST API.
func (s *Server) callLibreTranslate(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	body, err := json.Marshal(libreTranslateRequest{
		Q:      text,
		Source: sourceLang,
		Target: targetLang,
	})
	if err != nil {
		return "", status.Errorf(codes.Internal, "marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.libreTranslateAddr+"/translate", bytes.NewReader(body))
	if err != nil {
		return "", status.Errorf(codes.Internal, "create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", status.Errorf(codes.Unavailable, "libretranslate unavailable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		var errResp libreTranslateError
		if jsonErr := json.NewDecoder(resp.Body).Decode(&errResp); jsonErr == nil && errResp.Error != "" {
			return "", status.Errorf(codes.InvalidArgument, "libretranslate: %s", errResp.Error)
		}
		return "", status.Errorf(codes.InvalidArgument, "bad request to libretranslate: status %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return "", status.Errorf(codes.Internal, "unexpected libretranslate status %d", resp.StatusCode)
	}

	var result libreTranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", status.Errorf(codes.Internal, "decode response: %v", err)
	}

	return result.TranslatedText, nil
}
