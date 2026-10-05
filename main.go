package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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

	// Docker starts the API and Postgres together, so wait for the
	// database instead of crashing on the first refused connection.
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

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) { h.Healthz(c.Writer, c.Request) })

	// Logged group, the Gin version of a Laravel middleware group:
	// everything registered here passes through GinLogging.
	api := r.Group("/", middleware.GinLogging("/healthz"))
	api.GET("/rates", func(c *gin.Context) { h.Rates(c.Writer, c.Request) })
	api.POST("/orders", func(c *gin.Context) { h.Orders(c.Writer, c.Request) })

	log.Printf("shipquote listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, middleware.Logging(r, middleware.SkippedPaths("/healthz"))))
}
