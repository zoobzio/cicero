-- +goose Up
CREATE TABLE sources (
    hash TEXT PRIMARY KEY,
    text TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sources_tenant_id ON sources(tenant_id);

-- +goose Down
DROP TABLE sources;
