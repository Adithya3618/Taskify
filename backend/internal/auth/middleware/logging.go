package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs incoming requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request
		log.Printf("[REQUEST] %s %s - Started at %s", r.Method, r.URL.Path, start.Format("2006-01-02 15:04:05"))

		// Call next handler
		next.ServeHTTP(w, r)

		// Log completion
		duration := time.Since(start)
		log.Printf("[RESPONSE] %s %s - Completed in %v", r.Method, r.URL.Path, duration)
	})
}
