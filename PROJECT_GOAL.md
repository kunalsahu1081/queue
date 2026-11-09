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

## Core semantics

- **Delay ("maturity").** Each queue has a default delay and each message can override it. A message is invisible until `sentAt + delay`.
- **Visibility timeout.** A received message stays hidden for N seconds. Ack deletes it. With no ack in time, it reappears for another consumer.
- **At-least-once delivery.** A message can be delivered more than once, so consumers must be idempotent.
- **Receive count and DLQ.** Each delivery increments a counter. After `maxReceives` the message moves to a dead-letter queue.
- **Long polling.** `Receive` blocks for up to a wait time until a message matures, instead of busy-looping.

The key idea is that every message has one `visibleAt` timestamp. The delay sets it at send time and Receive pushes it forward, so one structure ordered by `visibleAt` handles both delay and visibility timeout.

## Scope

| Phase | Contents | Why it's there |
| --- | --- | --- |
| v1: core | Send with delay, Receive with long polling, Ack, Nack, ChangeVisibility, receive count, DLQ, config | Shows concurrency and data-structure skill |
| v2: server | HTTP/JSON API (gRPC optional), multiple named queues, graceful shutdown, Prometheus metrics | Shows production service hygiene |
| v3: durability | Append-only WAL, fsync policy, crash recovery, log compaction | Turns the Postgres/WAL background into an interview story |
| Stretch | Hierarchical timing wheel vs heap, sharded queues, FIFO groups | A benchmark-driven design write-up |

**Out of scope:** clustering and replication, auth, and exactly-once delivery. The README should name these as deliberate non-goals.

## Consumption design

### Message lifecycle

1. **Delayed:** `visibleAt` is in the future, so Receive skips the message.
2. **Visible:** `now >= visibleAt`, so the next Receive can take it.
3. **In flight:** handed to one consumer. Set `visibleAt = now + VisibilityTimeout`, issue a new `ReceiptHandle`, and increment `ReceiveCount`.
4. From there, one of three things happens:
    - **Ack:** the message is deleted.
    - **Nack, or the timeout passes with no ack:** the message becomes visible again. No background job is needed, because `visibleAt` has simply passed.
    - **`ReceiveCount > MaxRetries`:** the message becomes FAILED, or moves to the DLQ, and is never handed out again.

Only COMPLETE and FAILED need a stored status. The other states come from `visibleAt` and the receipt handle.

### API

```go
type Delivery struct {
    MessageID     string
    Body          []byte
    ReceiptHandle string
    ReceiveCount  int
}

Receive(ctx, queueID string, max int, wait time.Duration) ([]Delivery, error)
Ack(queueID, receipt string) error                 // ErrStaleReceipt if not current
Nack(queueID, receipt string, retryIn time.Duration) error
ExtendVisibility(queueID, receipt string, d time.Duration) error
```

### Inside Receive

1. Lock the queue.
2. Take up to `max` messages where `now >= visibleAt` and the status isn't final.
3. For each one: if `ReceiveCount >= MaxRetries`, mark it FAILED and skip it. Otherwise set `visibleAt = now + timeout`, issue a new receipt, and increment `ReceiveCount`.
4. Unlock. If you have at least one message, or `wait == 0`, return.
5. Otherwise block until Send signals a notify channel, a timer fires at the earliest future `visibleAt`, or `ctx` or `wait` ends. Then go back to step 1.

**Ack** locks the queue, finds the message by receipt, and rejects the ack if the receipt isn't the current one. Otherwise it deletes the message.

**The client side** is a pool of worker goroutines, each looping Receive → handle → Ack or Nack. Long jobs call `ExtendVisibility` every `timeout/2`.

**Compared with pg-boss.** pg-boss polls Postgres every couple of seconds using `FOR UPDATE SKIP LOCKED`. Our mutex plays the role of SKIP LOCKED, `startAfter` maps to the delay, and `expireIn` maps to the visibility timeout. Because the queue is in memory, long polling can replace interval polling.

## The 12-step build plan

Each step ends with something you can commit and test. Do them in order.

### Phase 1: core library

1. **Read the SQS semantics before writing code.** Write a one-page note in your own words. It becomes the first section of the README.
    - Done when: you can explain why SQS is at-least-once, and what happens when a visibility timeout expires mid-processing.
    - Read: [SQS delay queues](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-delay-queues.html), [Visibility timeout](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-visibility-timeout.html), [Dead-letter queues](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-dead-letter-queues.html), [Short vs long polling](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-short-and-long-polling.html)
2. **Set up the module and the public API without any logic.** Define `Message`, `Queue`, `Config`, and the method signatures, plus typed errors (`ErrNotFound`, `ErrReceiptExpired`).
    - Done when: `go vet` passes and the godoc reads well.
    - Read: [Organizing a Go module](https://go.dev/doc/modules/layout), [Functional options](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis), [Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
3. **Build a single-threaded queue on a min-heap keyed by `visibleAt`.** Implement Send and a non-blocking Receive, and inject a clock so tests never sleep.
    - Done when: a message with a 5s delay is invisible at 4.9s and visible at 5s on a fake clock.
    - Read: [container/heap](https://pkg.go.dev/container/heap) (study the PriorityQueue example and its `index` field)
4. **Add receipt handles, Ack, Nack, and ChangeVisibility.** Issue a new handle on every receive so stale acks are rejected. Use `heap.Fix` when `visibleAt` changes.
    - Done when: tests cover an ack after the timeout has expired (rejected) and a double ack.
5. **Make it concurrent.** Put one `sync.Mutex` around the heap and run every test with `-race`. Add a stress test with 64 producers and 64 consumers that checks no message is lost or delivered twice while in flight.
    - Done when: `go test -race -count=50` is green.
    - Read: [Go race detector](https://go.dev/blog/race-detector), [Share memory by communicating](https://go.dev/blog/codelab-share)
6. **Add long-polling Receive.** `Receive(ctx, max, wait)` blocks until a message matures, the wait ends, or ctx is cancelled. Wake sleepers on Send and when the earliest `visibleAt` arrives, using one timer plus a notify channel. Write down why you didn't use `sync.Cond` (it can't select on ctx).
    - Done when: there are no goroutine leaks and cancellation returns promptly.
    - Read: [Context](https://go.dev/blog/context), [Pipelines and cancellation](https://go.dev/blog/pipelines), [Timing out, moving on](https://go.dev/blog/concurrency-timeouts), [Go 1.23 timers](https://go.dev/wiki/Go123Timer), [testing/synctest](https://go.dev/blog/synctest)
7. **Add the receive count and DLQ.** After `maxReceives`, move the message to a DLQ (another Queue), and add a redrive function.
    - Done when: a poison message ends up in the DLQ after exactly N receives.
8. **Benchmark and profile.** Measure throughput at 1, 8, and 64 consumers, and p99 receive latency. Try sharding (N heaps with N locks) and keep it only if the numbers justify it.
    - Done when: the README has a benchmark table and CPU and mutex profiles.
    - Read: [testing.B.Loop](https://go.dev/blog/testing-b-loop), [Profiling Go programs](https://go.dev/blog/pprof)

### Phase 2: server

9. **Add an HTTP server with named queues.** Endpoints are CreateQueue, Send, Receive (long poll), Ack, and ChangeVisibility, as JSON over `net/http`. Shut down gracefully on SIGTERM, with in-flight long polls returning empty.
    - Done when: integration tests start the server on a random port and run the step 5 stress test over the network.
    - Read: [Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown), optionally the [gRPC Go quickstart](https://grpc.io/docs/languages/go/quickstart/)
10. **Add observability.** Expose queue depth, in-flight, delayed, and receive-latency metrics in Prometheus format, and use structured logs with `log/slog`.
    - Read: [Structured logging with slog](https://go.dev/blog/slog)

### Phase 3: durability

11. **Add a write-ahead log.** Every Send, Ack, and visibility change becomes a length-prefixed, CRC-checked record, appended before the call returns. Make the fsync policy configurable (always, or every N ms) and benchmark both.
    - Done when: after killing the process mid-write, recovery rebuilds the exact state, and only a truncated tail record is dropped.
    - Read: [Write-ahead log pattern](https://martinfowler.com/articles/patterns-of-distributed-systems/write-ahead-log.html), [PostgreSQL WAL](https://www.postgresql.org/docs/current/wal-intro.html)
12. **Add snapshots and compaction.** Periodically write out the live state and truncate the log. Recovery then becomes the snapshot plus the log tail.
    - Stretch read: [Kafka purgatory and hierarchical timing wheels](https://www.confluent.io/blog/apache-kafka-purgatory-hierarchical-timing-wheels/)

**Progress so far:** the HTTP server and the queue and message CRUD from step 9 came early. The consumption steps (4 to 7) are next, and the owner is implementing them.

## README checklist

- [ ] A one-paragraph pitch and a 10-line usage example
- [ ] Semantics: at-least-once delivery, delay and visibility, and what a receipt handle protects against
- [ ] Design decisions and rejected alternatives: heap vs timing wheel, one lock vs shards, timer and channel vs `sync.Cond`
- [ ] A benchmark table with machine specs and the command to reproduce it
- [ ] Durability: the fsync trade-off with numbers
- [ ] Non-goals: clustering, exactly-once delivery, auth
- [ ] CI on GitHub Actions running `go test -race`, `go vet`, and `staticcheck`
- [ ] A blog post on the long-polling wakeup problem, linked from the resume

## Working agreement for AI assistants

- **The owner writes the core logic themself.** Don't implement consumption or concurrency unless explicitly asked.
- **Don't refactor or restyle the owner's code.** Don't rename functions or move them between files unless asked. Variable and field renames, import fixes, and folder moves are fine when needed.
- **Give one clear recommendation, not a list of options,** and explain the Go reason behind it (for example, that the package is the visibility boundary).
- **Run `go vet ./...` and `go build ./...` after changes,** and say exactly what changed.
