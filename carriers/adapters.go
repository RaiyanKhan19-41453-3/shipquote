package carriers

// PathaoAdapter models a base-plus-per-kg tariff.
type PathaoAdapter struct{}

func (PathaoAdapter) Name() string { return "pathao" }

func (PathaoAdapter) Quote(origin, destination string, weightKG float64) (Quote, error) {
	if err := validate(weightKG); err != nil {
		return Quote{}, err
	}
	return Quote{Carrier: "pathao", PriceBDT: 60 + 25*weightKG, ETADays: 2}, nil
}

// RedXAdapter models a cheaper base with a steeper per-kg rate.
type RedXAdapter struct{}

func (RedXAdapter) Name() string { return "redx" }

func (RedXAdapter) Quote(origin, destination string, weightKG float64) (Quote, error) {
	if err := validate(weightKG); err != nil {
		return Quote{}, err
	}
	return Quote{Carrier: "redx", PriceBDT: 55 + 30*weightKG, ETADays: 3}, nil
}

// SteadfastAdapter models the cheapest base with a flat mid range rate.
type SteadfastAdapter struct{}

func (SteadfastAdapter) Name() string { return "steadfast" }

func (SteadfastAdapter) Quote(origin, destination string, weightKG float64) (Quote, error) {
	if err := validate(weightKG); err != nil {
		return Quote{}, err
	}
	return Quote{Carrier: "steadfast", PriceBDT: 50 + 28*weightKG, ETADays: 4}, nil
}

// Cheapest returns the lowest priced quote. ok is false when empty,
// so callers handle "no carrier available" explicitly instead of a zero value.
func Cheapest(quotes []Quote) (best Quote, ok bool) {
	for _, q := range quotes {
		if !ok || q.PriceBDT < best.PriceBDT {
			best, ok = q, true
		}
	}
	return best, ok
}
