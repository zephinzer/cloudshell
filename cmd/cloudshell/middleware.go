package main

import (
	"cloudshell/internal/log"
	"net/http"
)

func addLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s - %s%s", r.RemoteAddr, r.Host, r.URL.String())
		next.ServeHTTP(w, r)
	})
}
