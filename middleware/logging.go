package middleware

import (
	"log"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("%v %v\n", r.Method, r.RequestURI)
		next.ServeHTTP(rw, r)
	})
}
