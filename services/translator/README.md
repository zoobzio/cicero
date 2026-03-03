# translator

Self-hosted translation sidecar powered by [LibreTranslate](https://github.com/LibreTranslate/LibreTranslate).

## Configuration

The service runs as a Docker container using the official `libretranslate/libretranslate` image. Configuration is managed via docker-compose environment variables.

| Variable | Value | Purpose |
|----------|-------|---------|
| `LT_LOAD_ONLY` | `en,es,fr,de,pt,it,nl,pl,ru,zh,ja,ko` | Languages to load at startup |
| `LT_DISABLE_WEB_UI` | `true` | Disable the web UI (API only) |

## API

**Endpoint:** `POST /translate`

**Request:**
```json
{
  "q": "Hello, world!",
  "source": "en",
  "target": "es"
}
```

**Response:**
```json
{
  "translatedText": "¡Hola, mundo!"
}
```

Language codes follow ISO 639-1 (e.g., `en`, `es`, `fr`, `de`).

## Running Locally

```bash
docker compose up translator
```

The service exposes port `5000` and includes a health check on `/languages`.
