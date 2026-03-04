# translator

Go gRPC translation service. Receives translation requests over gRPC, selects a backend provider based on the routing classification, and returns translated text with provider metadata.

## Architecture

The translator service is the abstraction boundary between cicero's main application and translation backends. The application communicates with the service over gRPC. The service owns provider selection, error normalization, and backend communication.

```
app (gRPC client) --> translator service (gRPC) --> LibreTranslate (HTTP)
```

The service currently supports one backend (LibreTranslate) for simple-routed requests. Complex-routed requests (LLM provider) return `Unimplemented` until a subsequent issue adds that capability.

## Proto Contract

Defined in `proto/translator/translator.proto`:

```protobuf
service TranslatorService {
  rpc Translate(TranslateRequest) returns (TranslateResponse);
}

message TranslateRequest {
  string text = 1;
  string source_language = 2;
  string target_language = 3;
  string route = 4;            // "simple" or "complex"
}

message TranslateResponse {
  string translated_text = 1;
  string provider = 2;         // backend that handled the request
}
```

## Service Responsibilities

| Responsibility | Description |
|----------------|-------------|
| Provider selection | Routes requests to the appropriate backend based on the `route` field |
| Error normalization | Maps backend-specific errors to gRPC status codes (`InvalidArgument`, `Unavailable`, `Internal`) |
| Provider metadata | Returns which backend handled the request in the `provider` response field |

## Configuration

Environment variables for the translator service container:

| Variable | Default | Purpose |
|----------|---------|---------|
| `TRANSLATOR_LISTEN_ADDR` | `:9091` | gRPC listen address |
| `LIBRETRANSLATE_ADDR` | `http://localhost:5000` | LibreTranslate REST API address |

## Docker Compose Topology

The service runs as a container built from `services/translator/Dockerfile`. LibreTranslate runs as a separate container. The translator depends on LibreTranslate being healthy before starting.

```
libretranslate (port 5000, REST)  <--  translator (port 9091, gRPC)  <--  app (port 8080)
```

The application connects to the translator at `translator:9091` over gRPC. The translator connects to LibreTranslate at `http://libretranslate:5000` over HTTP.

## Running Locally

Start the full stack:

```bash
docker compose up translator
```

This starts both `libretranslate` and `translator` (the translator depends on LibreTranslate's health check).

## Client

The gRPC client in `external/translator/client.go` wraps the translator service with a pipz resilience stack: circuit breaker, backoff retry, and timeout. The client satisfies the `contracts.Translator` interface and is resolved from the sum registry by the translation pipeline.

## Language Codes

Language codes follow ISO 639-1 (e.g., `en`, `es`, `fr`, `de`). Supported languages depend on the LibreTranslate backend configuration (`LT_LOAD_ONLY` in docker-compose).
