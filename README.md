# Notification Service

A simple HTTP service that receives notifications, validates them, and optionally forwards them (e.g. to Microsoft Teams).  
The service exposes health endpoints and Prometheus metrics.

---

## Features

- REST endpoint to receive notifications
- Validation of input (strict JSON, required fields)
- Forwarding of `Warning` notifications
- Support for:
    - in-memory sink (default)
    - Microsoft Teams webhook
- Structured logging (`slog`)
- Request ID middleware
- Panic recovery
- Prometheus metrics
- Health and readiness endpoints
- Docker support

---

## API

### POST `/notifications`

Accepts a notification:

```json
{
  "type": "Warning",
  "name": "Backup Failure",
  "description": "Database issue"
}
```

## Supported types

- `Warning` → forwarded to sink
- `Info` → accepted but not forwarded

## Response

```json
{
  "accepted": true,
  "forwarded": true
}
```

## Endpoints

### `GET /healthz`

```json
{
  "status": "ok"
}
```

### `GET /readyz`

```json
{
  "status": "ready"
}
```

### `GET /metrics`

Prometheus metrics endpoint.

## Run locally (Go)

```bash
go run ./cmd/notification-service
```

Service will be available at:

```text
http://localhost:8080
```

## Run with Docker

### Build

```bash
docker build -t notification-service .
```

### Run

```bash
docker run -p 8080:8080 notification-service
```

### Run with Teams integration

```bash
docker run -p 8080:8080 \
  -e TEAMS_WEBHOOK_URL="https://your-webhook-url" \
  notification-service
```

If no webhook is provided, the service uses an in-memory sink.

## Testing with curl

### Valid Warning (forwarded)

```bash
curl -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "Warning",
    "name": "Backup Failure",
    "description": "Database issue"
  }'
```

### Valid Info (not forwarded)

```bash
curl -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "Info",
    "name": "Daily Status",
    "description": "All systems operational"
  }'
```

### Invalid JSON

```bash
curl -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d 'invalid-json'
```

### Unsupported type

```bash
curl -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "Debug",
    "name": "Test",
    "description": "Not supported"
  }'
```

### Missing field

```bash
curl -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "Warning",
    "name": "Missing description"
  }'
```

### Health check

```bash
curl http://localhost:8080/healthz
```

### Readiness check

```bash
curl http://localhost:8080/readyz
```

### Metrics

```bash
curl http://localhost:8080/metrics
```

## Metrics

The service exposes Prometheus metrics, including:

- `notification_service_notifications_received_total`
- `notification_service_notifications_forwarded_total`
- `notification_service_notifications_rejected_total`
- `notification_service_notification_errors_total`
- `notification_service_http_requests_total`
- `notification_service_http_request_duration_seconds`

Additionally:

- Go runtime metrics (`go_*`)
- Process metrics (`process_*`)

## Middleware

The HTTP stack includes:

- Request ID (`X-Request-ID`)
- Structured logging
- Prometheus metrics
- Panic recovery

## Notes

- JSON is strictly validated:
    - unknown fields are rejected
    - only one JSON object allowed
- Request body size is limited to **1 MB**
- Service runs as **non-root** in Docker

