package main

import (
	"net/http"

	"golang.org/x/time/rate"
)

func (cfg *apiConfig) ratelimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := rate.NewLimiter(10, 100)
		if !limiter.Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
