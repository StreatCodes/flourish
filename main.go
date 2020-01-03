package main

import (
	"crypto/tls"
	"log"

	"github.com/emersion/go-imap/server"
)

func main() {
	cert, err := tls.LoadX509KeyPair("publickey.cer", "private.key")
	if err != nil {
		log.Fatal(err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	be := NewBackend()

	// Create a new server
	s := server.New(be)
	s.Addr = "192.168.20.3:1993"

	s.TLSConfig = cfg

	log.Printf("Starting IMAP server at %s", s.Addr)
	if err := s.ListenAndServeTLS(); err != nil {
		log.Fatal(err)
	}
}
