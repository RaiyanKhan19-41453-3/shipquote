//go:build ignore

// Old stdlib-only boot path, kept for reference while learning.
// Run the real program from main.go instead.
package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"shipquote/carriers"
	"shipquote/handlers"
	"shipquote/middleware"
	"shipquote/store"
)

//go:embed migrations/schema.sql
var schema string

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	dsn := env("DATABASE_URL", "postgres://ship:ship@localhost:5432/shipquote?sslmode=disable")
	port := env("PORT", "8080")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for {
		if err := db.PingContext(ctx); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			log.Fatalf("db never came up: %v", ctx.Err())
		case <-time.After(time.Second):
		}
	}

	s := store.New(db)
	if err := s.Migrate(ctx, schema); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	h := handlers.New([]carriers.Adapter{carriers.PathaoAdapter{}, carriers.RedXAdapter{}, carriers.SteadfastAdapter{}}, s)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.Healthz)
	mux.HandleFunc("GET /rates", h.Rates)
	mux.HandleFunc("POST /orders", h.Orders)

	log.Printf("shipquote listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, middleware.Logging(mux)))
}
