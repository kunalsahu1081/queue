# letsgo: an SQS-style delay queue in Go

## What we're building

An in-memory message queue, written from scratch in Go and served over HTTP, that works like Amazon SQS:

- **Delayed delivery ("maturity").** Each message is sent with a delay and can't be consumed until `sentAt + delay` has passed.
- **Concurrent consumers.** Any number of clients or goroutines can consume from the same queue, and each message goes to exactly one consumer at a time.
- **Retries and failure.** A message that isn't acknowledged becomes visible again. After `ConsumptionRetries` attempts it's marked FAILED, and later it moves to a dead-letter queue.
- **At-least-once delivery.** Consumers must be idempotent. Exactly-once delivery isn't a goal.

## Why (the goal)

This is a portfolio project to show production-grade Go for backend and infra roles. It should demonstrate concurrency done correctly (mutexes, the race detector, no goroutine leaks), clear package boundaries, tests, benchmarks, and a README that explains the design trade-offs. The owner has a strong Postgres background, so the later durability phase (a write-ahead log, fsync policy, crash recovery) is meant to become a strong interview story.

## Current state (Nov 2026)

- Module `letsgo`, Go 1.25.6, one dependency: `dario.cat/mergo`.
- An HTTP server on `:8080` (`cmd/app/main.go`) using `net/http` `ServeMux`.
- **Queue and message CRUD work, and have been tested end to end:**

| Route | Action |
| --- | --- |
| `GET /queues`, `POST /queues` | List queues, create a queue |
| `GET`, `PATCH`, `DELETE /queues/{id}` | Get, update, or delete a queue (deleting also removes its messages) |
| `GET`, `POST /queues/{id}/messages` | List messages, add one with `{"data": ..., "delaySeconds": N}` |
| `GET`, `PATCH`, `DELETE /queues/{id}/messages/{messageId}` | Get, update, or delete one message |

- **Consumption isn't built yet.** `GetMaturedMessages`, `MarkConsumptionFailed`, and `MarkMessageConsumed` exist in `entities/message.go` but only compile. The owner is implementing them, along with all concurrency.

## Structure

```
cmd/app/main.go                     server startup
internal/
  domain/
    entities/queue.go               Queue struct + queue CRUD (in-memory maps/slices)
    entities/message.go             QueueMessage + message CRUD + consumption stubs
    errors/errors.go                package qerrors: AppError, NotFound, UpdateMergeFailed
    interfaces/queue.go             QueueActions interface (unused so far)
  services/                         package service: request structs, calls entities
  transport/http/
    router.go, routes/, handlers/   HTTP layer; writeError maps NOT_FOUND to 404
  utilities/                        package utility: GenerateUnique, DecodeRequest, normalizers
```

The layers are transport → service → entities. State lives in package-level vars in `entities` (`Queues`, `QueueIdMap`, `MessageIdMap`).

**Field visibility is deliberate.** Config fields are exported (`ID`, `Name`, `ConsumptionRetries`, ...). Internal state is private, so only `entities` can change it: `messages`, `visibleAt`, `consumptionStatus`, `retryCount`. Callers pass a delay, never a timestamp.

## Roadmap

1. **Consumption (owner implementing):**
    - Receive returns a matured message and hides it for a visibility timeout (`visibleAt = now + timeout`).
    - Every receive issues a fresh receipt handle.
    - Ack and Nack take the receipt handle, so a stale consumer's ack is rejected.
    - Messages are redelivered automatically when the timeout passes.
2. **Concurrency:** a `sync.Mutex` around queue state, plus a test with 50 goroutines and 1000 messages where each is received exactly once under `go test -race`.
3. **Long polling:** `Receive(ctx, queueID, max, wait)` blocks until a message matures, woken by a notify channel on Send and a timer for the earliest `visibleAt`. This removes the need for pg-boss-style interval polling.
4. **Dead-letter queue** after max retries, plus redrive.
5. **Heap ordered by `visibleAt`** to replace the linear scan, with benchmarks at 1, 8, and 64 consumers.
6. **Durability:** an append-only WAL (length-prefixed, CRC), configurable fsync, crash recovery, snapshots and compaction.
7. **Observability:** Prometheus metrics (depth, in-flight, delayed, latency) and `log/slog`.
8. **README:** semantics, design decisions with the rejected alternatives, benchmark table, non-goals (clustering, exactly-once, auth).

## Working agreement for AI assistants

- **The owner writes the core logic themself.** Don't implement consumption or concurrency unless explicitly asked.
- **Don't refactor or restyle the owner's code.** Don't rename functions or move them between files unless asked. Variable and field renames, import fixes, and folder moves are fine when needed.
- **Give one clear recommendation, not a list of options,** and explain the Go reason behind it (for example, that the package is the visibility boundary).
- **Run `go vet ./...` and `go build ./...` after changes,** and say exactly what changed.
