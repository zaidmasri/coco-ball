// Package middleware has http middleware configurations
package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %s", r.Method, clientIP(r), r.URL.Path, time.Since(start))
	})
}

// clientIP returns the originating client address. Requests arrive through a
// Cloudflare Tunnel, so r.RemoteAddr is always the local cloudflared
// connection; the real visitor IP is carried in the CF-Connecting-IP header
// that Cloudflare's edge sets on every request it forwards.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if ip, _, ok := strings.Cut(fwd, ","); ok {
			return strings.TrimSpace(ip)
		}
		return strings.TrimSpace(fwd)
	}
	return r.RemoteAddr
}
