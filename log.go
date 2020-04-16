package main

import (
	"log"
	"net/http"
)

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr + " - " + r.Host + " - " + r.Method + " - " + r.URL.String() + " - " + r.Header.Get("User-Agent"))
		next.ServeHTTP(w, r)
	})
}