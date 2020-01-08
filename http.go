package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

func initHTTP(addr string, port int, certFile string, keyFile string) {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	listenAddr := fmt.Sprintf("%s:%d", addr, port)
	log.Printf("Starting HTTPS server at %s", listenAddr)
	http.ListenAndServeTLS(listenAddr, certFile, keyFile, r)
}
