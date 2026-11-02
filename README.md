# Stampede Shield

Stampede Shield is a small-scale reproduction of the cache-stampede problem class: when many requests miss a cache at once, each request can repeat the same expensive database read. It is not a reproduction of any company's actual system. Phase 1 provides the intentionally naive Redis-to-Postgres path so the failure can be measured before hardening.

## Run

```sh
docker compose up
```

The API is at `http://localhost:8080`, the dashboard at `http://localhost:5173`, and Prometheus metrics at `http://localhost:8080/metrics`.

Run the baseline load test with k6:

```sh
k6 run loadtest/burst.js
```

## Architecture

```mermaid
flowchart LR
  Browser --> Vue
  Vue --> API[Go API / chi]
  API --> Redis
  API --> Postgres[(Postgres)]
  API --> Prometheus[/metrics]
```

## Benchmark

No benchmark numbers are reported yet because Docker and k6 were unavailable in the development environment. Numbers must be recorded from actual k6 output, along with machine specifications.

| Scenario | Naive | Hardened |
|---|---:|---:|
| P99 latency under burst | pending Phase 1 run | pending Phase 2 |
| DB QPS during cold start | pending Phase 1 run | pending Phase 2 |
| Recovery after cache flush | pending Phase 1 run | pending Phase 2 |
| Error rate during Redis outage | pending Phase 1 run | pending Phase 4 |

## Roadmap

## Phase 2 hardening

The runtime `POST /mode?value=naive|hardened` endpoint switches between the baseline and hardened read paths. Hardened mode uses:

- `singleflight` with a detached 300 ms database context, so one cancelled client cannot cancel the shared read.
- Soft/hard TTLs with up to 20% jitter.
- Stale-while-revalidate, with a separate singleflight refresh key per feature.
- Database fallback when Redis is unavailable; stale values remain immediately usable while refresh fails.

The coalescing proof and unit tests cover 1,000 concurrent requests, jitter bounds, one refresh under concurrent stale reads, and Redis-down fallback:

```sh
cd backend
go test ./...
go test -race ./...
```

Later phases add SSE metrics, chaos controls, load shedding, and the complete dashboard.

## Phase 3 observability and dashboard

The dashboard consumes `GET /metrics/stream`, an SSE endpoint that emits JSON snapshots every 500 ms. It shows rolling P99 latency and database QPS charts, hit rate, coalescing/stale counters, connection state, and the active read mode. The built-in burst control calls:

```sh
curl -X POST 'http://localhost:8080/loadgen/burst?rps=200&seconds=10&keys=10'
```

The frontend is Vue 3 + Pinia + uPlot. Build it locally with:

```sh
cd frontend
npm install
npm run build
```

## Phase 4 resilience and chaos controls

The API now supports controlled failure injection:

```sh
curl -X POST http://localhost:8080/chaos/flush-cache
curl -X POST 'http://localhost:8080/chaos/redis-down?seconds=15'
curl -X POST 'http://localhost:8080/chaos/db-latency?ms=250'
curl http://localhost:8080/chaos/state
```

Database reads are capped at 100 in-flight operations. Excess requests receive `429` and increment `stampede_shed_requests_total`. Hardened mode serves stale cache entries immediately while refreshes fail or the database is slow; naive mode exposes the baseline failure behavior.

## Demo

`[Demo GIF placeholder]`
