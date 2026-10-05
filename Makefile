GO_IMG := golang:1.23-bookworm
GO := docker run --rm -e DATABASE_URL -e PORT -v .:/app -w /app $(GO_IMG) go

.PHONY: run test vet fmt build up down

run:
	docker run --rm -e DATABASE_URL -e PORT -p 8080:8080 -v .:/app -w /app $(GO_IMG) go run .

test:
	$(GO) test $(or $(TESTARGS),./...)

vet:
	$(GO) vet ./...

fmt:
	docker run --rm -v .:/app -w /app $(GO_IMG) gofmt -l .

build:
	$(GO) build -o /tmp/shipquote-build .

up:
	docker compose up --build -d

down:
	docker compose down
