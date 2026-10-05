package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Order is one booked shipment.
type Order struct {
	ID             int64     `json:"id"`
	IdempotencyKey string    `json:"idempotency_key"`
	Origin         string    `json:"origin"`
	Destination    string    `json:"destination"`
	WeightKG       float64   `json:"weight_kg"`
	Carrier        string    `json:"carrier"`
	PriceBDT       float64   `json:"price_bdt"`
	CreatedAt      time.Time `json:"created_at"`
}

// Store wraps the order table. Keep SQL here so handlers stay thin.
type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Migrate(ctx context.Context, schema string) error {
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// Create inserts the order, or returns the existing row when the same
// idempotency key is retried. created=false means "duplicate, no charge".
func (s *Store) Create(ctx context.Context, o Order) (Order, bool, error) {
	var out Order
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO orders (idempotency_key, origin, destination, weight_kg, carrier, price_bdt)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING id, idempotency_key, origin, destination, weight_kg, carrier, price_bdt, created_at`,
		o.IdempotencyKey, o.Origin, o.Destination, o.WeightKG, o.Carrier, o.PriceBDT,
	).Scan(&out.ID, &out.IdempotencyKey, &out.Origin, &out.Destination, &out.WeightKG, &out.Carrier, &out.PriceBDT, &out.CreatedAt)
	if err == nil {
		return out, true, nil
	}
	if err != sql.ErrNoRows {
		return Order{}, false, fmt.Errorf("create order: %w", err)
	}
	err = s.db.QueryRowContext(ctx, `
		SELECT id, idempotency_key, origin, destination, weight_kg, carrier, price_bdt, created_at
		FROM orders WHERE idempotency_key = $1`, o.IdempotencyKey,
	).Scan(&out.ID, &out.IdempotencyKey, &out.Origin, &out.Destination, &out.WeightKG, &out.Carrier, &out.PriceBDT, &out.CreatedAt)
	if err != nil {
		return Order{}, false, fmt.Errorf("fetch duplicate order: %w", err)
	}
	return out, false, nil
}
