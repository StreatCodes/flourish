package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/emersion/go-smtp"
)

//Config is the global application config
type Config struct {
	MailDir     string
	TLSCertFile string
	TLSKeyFile  string
}

var config Config

func initIMAP(address string, port int, tlsConfig *tls.Config) {
	// be := NewBackend()

	// s := server.New(be)
	// s.Addr = fmt.Sprintf("%s:%d", address, port)

	// s.TLSConfig = tlsConfig
	// log.Printf("Starting IMAP server at %s", s.Addr)
	// if err := s.ListenAndServeTLS(); err != nil {
	// 	log.Fatal(err)
	// }
}

func initSMTP(addr string, port int, tlsConfig *tls.Config) {
	address := fmt.Sprintf("%s:%d", addr, port)

	be := &SMTPBackend{}

	s := smtp.NewServer(be)

	s.Addr = address
	s.Domain = "xps-mort"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = false
	s.TLSConfig = tlsConfig

	log.Println("Starting SMTP server at", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	//Load config
	f, err := os.Open("./config.json")
	if err != nil {
		log.Fatal("No config.json found")
	}
	dec := json.NewDecoder(f)
	dec.Decode(&config)

	domains := LoadDomains(config.MailDir)
	fmt.Printf("%v", domains)

	//Load tls certs
	cert, err := tls.LoadX509KeyPair(config.TLSCertFile, config.TLSKeyFile)
	if err != nil {
		log.Fatal(err)
	}
	TLSConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	//Start servers
	go initHTTP("0.0.0.0", 8080, config.TLSCertFile, config.TLSKeyFile)
	go initIMAP("0.0.0.0", 2882, TLSConfig)
	go initSMTP("0.0.0.0", 2525, TLSConfig)

	//Wait forever
	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
