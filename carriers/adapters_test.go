package carriers

import "testing"

func TestCheapestPicksMinimum(t *testing.T) {
	quotes := []Quote{
		{Carrier: "pathao", PriceBDT: 110, ETADays: 2},
		{Carrier: "redx", PriceBDT: 100, ETADays: 3},
	}
	best, ok := Cheapest(quotes)
	if !ok || best.Carrier != "redx" {
		t.Fatalf("got %+v, %v; want redx, true", best, ok)
	}
}

func TestCheapestEmpty(t *testing.T) {
	if _, ok := Cheapest(nil); ok {
		t.Fatal("want ok=false for no quotes")
	}
}

func TestAdaptersRejectBadWeight(t *testing.T) {
	for _, a := range []Adapter{PathaoAdapter{}, RedXAdapter{}} {
		if _, err := a.Quote("DHK", "CTG", 0); err == nil {
			t.Fatalf("%s accepted zero weight", a.Name())
		}
		if _, err := a.Quote("DHK", "CTG", -2); err == nil {
			t.Fatalf("%s accepted negative weight", a.Name())
		}
	}
}

func TestAdaptersPriceMath(t *testing.T) {
	q, err := PathaoAdapter{}.Quote("DHK", "CTG", 2)
	if err != nil || q.PriceBDT != 110 {
		t.Fatalf("pathao 2kg: got %+v, %v; want 110", q, err)
	}
	q, err = RedXAdapter{}.Quote("DHK", "CTG", 2)
	if err != nil || q.PriceBDT != 115 {
		t.Fatalf("redx 2kg: got %+v, %v; want 115", q, err)
	}
	q, err = SteadfastAdapter{}.Quote("DHK", "CTG", 2)
	if err != nil || q.PriceBDT != 106 {
		t.Fatalf("steadfast 2kg: got %+v, %v; want 106", q, err)
	}
}
