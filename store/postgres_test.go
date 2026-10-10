package store

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// TestCreateIsIdempotent needs a database. Run it with:
//
//	DATABASE_URL=postgres://ship:ship@localhost:5432/shipquote?sslmode=disable go test ./store/ -run TestCreateIsIdempotent -v
func TestCreateIsIdempotent(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL unset, skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	s := New(db)
	// The test owns its schema: CI hands us a blank database, so create
	// the table here instead of assuming the server already migrated it.
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		idempotency_key TEXT UNIQUE NOT NULL,
		origin TEXT NOT NULL,
		destination TEXT NOT NULL,
		weight_kg DOUBLE PRECISION NOT NULL,
		carrier TEXT NOT NULL,
		price_bdt DOUBLE PRECISION NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE orders`); err != nil {
		t.Fatal(err)
	}
	in := Order{IdempotencyKey: "key-1", Origin: "DHK", Destination: "CTG", WeightKG: 2, Carrier: "pathao", PriceBDT: 110}
	first, created, err := s.Create(ctx, in)
	if err != nil || !created {
		t.Fatalf("first insert: %+v %v %v", first, created, err)
	}
	second, created, err := s.Create(ctx, in)
	if err != nil || created {
		t.Fatalf("retry: %+v %v %v", second, created, err)
	}
	if first.ID != second.ID {
		t.Fatalf("retry returned different row: %d vs %d", first.ID, second.ID)
	}
}
