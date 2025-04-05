package main

import (
	"net/http"
)

func (cfg *apiConfig) validateIp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedIpAdd := "100.28.201.155"
		clientIp := r.RemoteAddr
		if forwarded := r.Header.Get("x-Forwarded-For"); forwarded != "" {
			clientIp = forwarded
		}
		if clientIp != allowedIpAdd {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
