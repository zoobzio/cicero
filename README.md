# Cicero

Cost-intelligent translation pipeline for the zoobz.io ecosystem.

## Overview

Cicero translates text content by classifying complexity and routing to the appropriate provider. Straightforward content goes to a self-hosted translation sidecar. Nuanced content escalates to an LLM. Translations are stored and deduplicated by content hash.

- **Content-addressable storage** — source text is hashed (SHA-256, truncated to 32 hex chars). The hash is returned to the caller and used for retrieval. Same text from different callers produces the same hash.
- **Dual-provider routing** — a complexity classifier determines whether text routes to the sidecar (bulk, infrastructure cost) or an LLM (nuanced, per-request cost).
- **Multi-tenancy** — the API supports multiple customers without structural changes.
- **Explicit language specification** — the caller specifies source and target language.

## Project Structure

```
cmd/
├── app/              # Public API binary
└── admin/            # Admin API binary

# Shared layers
config/               # Configuration types
models/               # Domain models
stores/               # Data access (shared, satisfies multiple contracts)
events/               # Domain events and signals
migrations/           # Database migrations (goose)
internal/             # Internal packages (pipeline, classifier)
external/             # External service clients
testing/              # Test infrastructure, mocks, fixtures

# Public API surface
api/
├── contracts/        # Public interface definitions
├── wire/             # Public request/response types
├── handlers/         # Public HTTP handlers
└── transformers/     # Public model <-> wire mapping

# Admin API surface
admin/
├── contracts/        # Admin interface definitions
├── wire/             # Admin request/response types
├── handlers/         # Admin HTTP handlers
└── transformers/     # Admin model <-> wire mapping

# Proto definitions (separate Go submodule)
proto/
└── translator/       # TranslatorService gRPC contract

# Translation sidecar (separate Go submodule)
services/
└── translator/       # Go gRPC server wrapping LibreTranslate

# Sidecar client
external/
└── translator/       # gRPC client with pipz resilience
```

## Getting Started

```bash
# Install dependencies
go mod tidy

# Run the application
make run

# Run tests
make test

# Run linter
make lint

# Full CI check
make check
```

## Development

### Prerequisites

- Go 1.25+
- golangci-lint v2.7.2
- protoc with protoc-gen-go and protoc-gen-go-grpc (for proto changes)

### Install Tools

```bash
make install-tools
make install-hooks
```

### Make Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the application binary |
| `make run` | Run the application |
| `make test` | Run all tests with race detector |
| `make test-unit` | Run unit tests only |
| `make test-integration` | Run integration tests |
| `make test-bench` | Run benchmarks |
| `make lint` | Run linters |
| `make coverage` | Generate coverage report |
| `make check` | Run tests + lint |
| `make ci` | Full CI simulation |
| `make dev` | Start development environment (docker compose) |
| `make dev-down` | Stop development environment |
| `make dev-logs` | Tail application logs |
| `make dev-reset` | Reset development environment (removes volumes) |

## Architecture

Cicero uses a dual API surface architecture with shared domain layers:

**Shared layers** (used by all surfaces):
- **models** — Domain models, no internal dependencies
- **stores** — Data access implementations (same store satisfies multiple contracts)
- **migrations** — Database schema
- **events** — Domain events and signals
- **config** — Configuration types

**Surface-specific layers** (each surface has its own):
- **contracts** — Interface definitions
- **wire** — Request/response types
- **handlers** — HTTP layer
- **transformers** — Pure mapping functions between models and wire

**Internal packages:**
- **internal/classify** — Text complexity classification
- **internal/translate** — Translation pipeline (hash, deduplicate, classify, translate, store)

**Proto definitions:**
- **proto/translator** — gRPC service contract for the translation sidecar (separate Go submodule)

**Services:**
- **services/translator** — Go gRPC server that wraps LibreTranslate, owns provider selection and error normalization (separate Go submodule)

**External clients:**
- **external/translator** — gRPC client with pipz resilience stack (circuit breaker, backoff, timeout)

## License

MIT
