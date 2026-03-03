# Mission: cicero

Cost-intelligent translation pipeline for the zoobz.io ecosystem.

## Purpose

Translate text content by classifying complexity and routing to the appropriate provider. Straightforward content goes to a self-hosted translation service. Nuanced content escalates to an LLM. Translations are stored and deduplicated by content hash.

## Core Concepts

### Complexity Classification

A deterministic algorithm evaluates incoming text against criteria and thresholds. Certain signals trigger automatic LLM escalation (e.g., markdown content). Others are evaluated quantitatively (content length, structural markers). The classifier produces a routing decision.

### Dual-Provider Routing

Self-hosted sidecar for bulk translations (infra cost). LLM for nuanced content.

### Content-Addressable Storage

Source text is hashed. The hash is returned to the caller and used for retrieval. One source record, many translation records — each for a different target language. Same source text from different callers produces the same hash. Existing translations are served, not re-translated.

### Multi-Tenancy

zoobz.io is the first customer. The API supports additional customers without structural changes.

### Explicit Language Specification

The caller specifies source and target language.

## Translation Sidecar

A self-hosted translation service running as a sidecar container. Part of this repository. Follows the same sidecar pattern established in vicky: separate service in `services/`, gRPC communication, client with resilience patterns in `external/`, docker-compose for local development.

## The Stack

| Package | Purpose |
|---------|---------|
| `sum` | Service registry, dependency injection, boundaries |
| `rocco` | HTTP handlers, OpenAPI generation, SSE streaming |
| `grub` | Storage abstraction (Database, Bucket, Store, Index) |
| `soy` | Type-safe SQL query builder |
| `pipz` | Composable pipeline workflows |
| `flux` | Hot-reload runtime configuration (capacitors) |
| `cereal` | Field-level encryption, hashing, masking |
| `capitan` | Events and observability signals |
| `check` | Request validation |

## API Surfaces

| Surface | Binary | Consumer | Characteristics |
|---------|--------|----------|-----------------|
| `api` | `cmd/app/` | Customers | Tenant-scoped, submit translations, retrieve by hash |
| `admin` | `cmd/admin/` | Internal team | Cross-tenant visibility, provider management, usage metrics |

### Layer Organization

**Shared layers** (used by all surfaces):
- `models/` — Domain models
- `stores/` — Data access implementations (same store satisfies multiple contracts)
- `migrations/` — Database schema
- `events/` — Domain events
- `config/` — Configuration

**Surface-specific layers** (each surface has its own):
- `{surface}/contracts/` — Interface definitions
- `{surface}/wire/` — Request/response types (different masking per surface)
- `{surface}/handlers/` — HTTP handlers
- `{surface}/transformers/` — Model <-> wire mapping

### Surface Differences

| Aspect | Public (api/) | Admin (admin/) |
|--------|---------------|----------------|
| Auth | Customer identity | Admin/internal identity |
| Scope | Tenant's own data | System-wide access |
| Operations | Submit translations, retrieve by hash | Bulk ops, provider management, audit |
| Data exposure | Masked, minimal | Full visibility |

Note: Stores are shared. The same store implementation satisfies both `api/contracts` and `admin/contracts` interfaces.

## Project Structure

```
cmd/
├── app/              # Public API binary
└── admin/            # Admin API binary

# Shared layers
config/               # Static configuration types
models/               # Domain models
stores/               # Data access (shared, satisfies multiple contracts)
events/               # Domain events and signals
migrations/           # Database migrations (goose)
internal/             # Internal packages
testing/              # Test infrastructure, mocks, fixtures

# Public API surface
api/
├── contracts/        # Public interface definitions
├── wire/             # Public request/response types (masked)
├── handlers/         # Public HTTP handlers
└── transformers/     # Public model <-> wire mapping

# Admin API surface
admin/
├── contracts/        # Admin interface definitions
├── wire/             # Admin request/response types (unmasked)
├── handlers/         # Admin HTTP handlers
└── transformers/     # Admin model <-> wire mapping

# Translation sidecar
services/
└── translator/       # Self-hosted translation service

# Sidecar client
external/
└── translator/       # gRPC client with resilience patterns
```

## Conventions

### Naming

| Layer | File | Type | Example |
|-------|------|------|---------|
| Model | `models/source.go` | `Source` (singular) | `type Source struct` |
| Store | `stores/sources.go` | `Sources` (plural struct) | `type Sources struct` |
| Contract | `{surface}/contracts/sources.go` | `Sources` (plural interface) | `type Sources interface` |
| Wire | `{surface}/wire/sources.go` | Singular + suffix | `SourceResponse` |
| Handler | `{surface}/handlers/sources.go` | Verb+Singular | `var GetSource`, `var CreateSource` |

### Registration Points

After creating artifacts, wire them appropriately:

**Shared:**
- `stores/stores.go` — aggregate factory (all stores)
- `models/boundary.go` — model boundaries

**Surface-specific (replace `{surface}` with `api` or `admin`):**
- `{surface}/handlers/handlers.go` — `All()` function
- `{surface}/handlers/errors.go` — domain errors
- `{surface}/wire/boundary.go` — wire boundaries (masking for public API)

### Testing

- 1:1 relationship: `source.go` -> `source_test.go`
- Helpers in `testing/` call `t.Helper()`
- Mocks use function-field pattern
- Fixtures return test data with sensible defaults

## Non-Goals

- Language detection or auto-identification
- Translation memory or fuzzy matching
- Training or fine-tuning translation models
