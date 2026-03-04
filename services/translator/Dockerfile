# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git ca-certificates

# Copy proto submodule (required by replace directive)
COPY proto/ proto/

# Copy translator service
COPY services/translator/ services/translator/

# Build binary
WORKDIR /build/services/translator
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /translator ./cmd/translator

# Runtime stage
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /translator /app/translator

RUN adduser -D -g '' translator
USER translator

EXPOSE 9091

ENTRYPOINT ["/app/translator"]
