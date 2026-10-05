# shipquote

Shipping rate aggregator API in Go. It normalizes quotes from multiple
carriers behind one contract, books against the cheapest, and makes
order creation idempotent so retries never double charge.

Built as a focused Go learning project in the shape of real transaction
backend work: services, PostgreSQL, Docker Compose, and tests.

## Prerequisites

Docker and Docker Compose only. No local Go needed, everything runs
through the `golang:1.23-bookworm` image.

## Run it

```bash
docker compose up --build -d   # .env ships with working defaults
curl localhost:8080/healthz   # {"status":"ok"}
```

## Try it

Compare carriers for a 2kg Dhaka to Chittagong parcel:

```bash
curl "localhost:8080/rates?from=DHK&to=CTG&weight=2"
```

```json
{"cheapest":{"carrier":"steadfast","price_bdt":106,"eta_days":4},"quotes":[...]}
```

Book the cheapest option. Safe to retry, same key means same order:

```bash
curl -X POST localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"idempotency_key":"order-1","from":"DHK","to":"CTG","weight":2}'
# first time: 201 with "duplicate": false
# exact same request again: 200 with "duplicate": true, same row
```

## Endpoints

| Method | Path      | Success            | Errors                     |
|--------|-----------|--------------------|----------------------------|
| GET    | /healthz  | 200 `{"status":"ok"}` |                         |
| GET    | /rates    | 200 quotes plus cheapest | 400 bad params, 502 carrier down, 503 none available |
| POST   | /orders   | 201 new, 200 retry | 400 bad body, 502 carrier down, 503 none available, 500 save failed |

## Layout

- `carriers/` — one interface per carrier, normalized quotes, cheapest pick, unit tests
- `store/` — PostgreSQL orders with idempotent insert, plus a DB gated integration test
- `handlers/` — thin HTTP layer: parse, validate, delegate, respond
- `middleware/` — request logging with status and latency
- `migrations/schema.sql` — embedded and applied at startup
- `main.go` — boot: config, database wait, migrate, Gin routes around the handlers
- `old_main.go` — the original stdlib-only boot, parked with a build ignore tag for reference

## Test it

```bash
make test                                                # unit tests, no DB needed
DATABASE_URL=... make test TESTARGS='./store/ -run TestCreateIsIdempotent -v'
```

## Why it is built this way

- One carrier contract so new carriers never touch existing code
- Idempotency keys plus a unique database constraint so retries are safe
- stdlib core with an optional Gin serving layer, matching what most BD teams use
