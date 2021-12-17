package middleware

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		now := time.Now()
		next.ServeHTTP(rw, r)
		duration := time.Since(now)
		log.Printf("%v %v %vms %v\n", r.Method, r.RequestURI, duration.Milliseconds(), r.RemoteAddr)
	})
}
