# Semiconductor Excursion Containment Service

This service coordinates containment decisions after semiconductor manufacturing excursions. It keeps wafer lots isolated while evidence is collected, propagates request cancellation through equipment sampling, validates recipe policy, snapshots wafer selections without aliasing, manages evidence resources, serves concurrency-safe metrology summaries, advances containment states, classifies supplier evidence errors, fans out chamber investigations, and preserves review deadlines.

## Layout

- `cmd/containment`: HTTP service entry point.
- `internal/containment`: lot containment state and transitions.
- `internal/quarantine`: quarantine persistence and error classification.
- `internal/sampling`: cancellable equipment sampling.
- `internal/recipe`: recipe policy loading and validation.
- `internal/wafer`: immutable wafer selection snapshots.
- `internal/evidence`: evidence resource and transaction lifecycle.
- `internal/metrology`: concurrent measurement registry.
- `internal/supplier`: supplier evidence retry classification.
- `internal/investigation`: bounded chamber investigation fan-out.
- `internal/review`: deadline-aware engineering review.
- `web`: small operator status page.

## Run

```bash
go run ./cmd/containment
```

The server listens on `HTTP_ADDR` (default `:8080`). `GET /healthz` returns service health and `GET /api/v1/summary` returns a containment capability summary.

## Test

```bash
go test ./...
go test -race ./...
```

## Environment

- `HTTP_ADDR`: HTTP listen address.
- `SERVICE_NAME`: health response service name.
- `REVIEW_TIMEOUT`: default engineering review timeout such as `2s`.
