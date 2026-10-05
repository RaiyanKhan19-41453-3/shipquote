package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"shipquote/carriers"
	"shipquote/store"
)

type Handler struct {
	adapters []carriers.Adapter
	store    *store.Store
}

func New(adapters []carriers.Adapter, s *store.Store) *Handler {
	return &Handler{adapters: adapters, store: s}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// Healthz is the liveness probe load balancers and compose healthchecks hit.
func (h *Handler) Healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Rates normalizes quotes from every carrier and returns the cheapest.
// GET /rates?from=DHK&to=CTG&weight=2
func (h *Handler) Rates(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from, to := q.Get("from"), q.Get("to")
	weight, err := strconv.ParseFloat(q.Get("weight"), 64)
	if from == "" || to == "" || err != nil || weight <= 0 {
		writeErr(w, http.StatusBadRequest, "need ?from=DHK&to=CTG&weight=2 (weight positive)")
		return
	}
	quotes := make([]carriers.Quote, 0, len(h.adapters))
	for _, a := range h.adapters {
		quote, err := a.Quote(from, to, weight)
		if err != nil {
			writeErr(w, http.StatusBadGateway, "carrier "+a.Name()+" failed: "+err.Error())
			return
		}
		quotes = append(quotes, quote)
	}
	best, ok := carriers.Cheapest(quotes)
	if !ok {
		writeErr(w, http.StatusServiceUnavailable, "no carrier available")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"quotes": quotes, "cheapest": best})
}

type orderRequest struct {
	IdempotencyKey string  `json:"idempotency_key"`
	From           string  `json:"from"`
	To             string  `json:"to"`
	Weight         float64 `json:"weight"`
}

// Orders books against the cheapest carrier. Safe to retry: the same
// idempotency key returns the original row with duplicate=true and 200.
func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {
	var req orderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.IdempotencyKey == "" || req.From == "" || req.To == "" || req.Weight <= 0 {
		writeErr(w, http.StatusBadRequest, "need idempotency_key, from, to, positive weight")
		return
	}
	quotes := make([]carriers.Quote, 0, len(h.adapters))
	for _, a := range h.adapters {
		quote, err := a.Quote(req.From, req.To, req.Weight)
		if err != nil {
			writeErr(w, http.StatusBadGateway, "carrier "+a.Name()+" failed: "+err.Error())
			return
		}
		quotes = append(quotes, quote)
	}
	best, ok := carriers.Cheapest(quotes)
	if !ok {
		writeErr(w, http.StatusServiceUnavailable, "no carrier available")
		return
	}
	order, created, err := h.store.Create(r.Context(), store.Order{
		IdempotencyKey: req.IdempotencyKey,
		Origin:         req.From,
		Destination:    req.To,
		WeightKG:       req.Weight,
		Carrier:        best.Carrier,
		PriceBDT:       best.PriceBDT,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not save order")
		return
	}
	if !created {
		writeJSON(w, http.StatusOK, map[string]any{"order": order, "duplicate": true})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"order": order, "duplicate": false})
}
