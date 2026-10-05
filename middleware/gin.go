package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// GinLogging is the group middleware: attach it to a route group and
// every route inside gets method, path, status, and latency logged,
// exactly like assigning middleware to a Laravel route group.
// Paths in skip pass through silently, e.g. health checks.
func GinLogging(skip ...string) gin.HandlerFunc {
	skipped := make(map[string]bool, len(skip))
	for _, p := range skip {
		skipped[p] = true
	}
	return func(c *gin.Context) {
		if skipped[c.FullPath()] {
			c.Next()
			return
		}
		start := time.Now()
		c.Next()
		log.Printf("%s %s %d %s", c.Request.Method, c.Request.URL.Path,
			c.Writer.Status(), time.Since(start).Round(time.Millisecond))
	}
}
