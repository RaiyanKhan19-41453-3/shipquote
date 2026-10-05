package middleware

import (
	"log"
	"net/http"
	"time"
)

type recorder struct {
	http.ResponseWriter
	status int
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Skipper decides which requests pass through unlogged.
// Health checks are the classic case: balancers hit them every few
// seconds and they would only noise up the logs.
type Skipper func(*http.Request) bool

// SkippedPaths builds a Skipper matching exact paths.
func SkippedPaths(paths ...string) Skipper {
	return func(r *http.Request) bool {
		for _, p := range paths {
			if r.URL.Path == p {
				return true
			}
		}
		return false
	}
}

// Logging records method, path, status, and latency per request,
// except those the Skipper excuses. A nil Skipper logs everything.
func Logging(next http.Handler, skip Skipper) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skip != nil && skip(r) {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rec := &recorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond))
	})
}
