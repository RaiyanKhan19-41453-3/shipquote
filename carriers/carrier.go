package carriers

import "fmt"

// Quote is a normalized shipping quote. Prices are always in BDT so
// callers never juggle per-carrier currencies or field names.
type Quote struct {
	Carrier  string  `json:"carrier"`
	PriceBDT float64 `json:"price_bdt"`
	ETADays  int     `json:"eta_days"`
}

// Adapter is one carrier behind a common contract. Real adapters would
// call the carrier HTTP API here; the fakes below model the pricing math.
type Adapter interface {
	Name() string
	Quote(origin, destination string, weightKG float64) (Quote, error)
}

func validate(weightKG float64) error {
	if weightKG <= 0 {
		return fmt.Errorf("weight_kg must be positive, got %v", weightKG)
	}
	return nil
}
