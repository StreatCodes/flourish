package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"golang.org/x/crypto/bcrypt"
)

//AdminSession is used for authentication with the HTTP API
type AdminSession struct {
	Username string
	Token    []byte
}

//AdminSessions is a helper struct to keep track and retrieve AdminSession's
type AdminSessions struct {
	sessions []AdminSession
}

var sessions AdminSessions

//Get returns the AdminSession and true if it exists
func (a *AdminSessions) Get(token []byte) (AdminSession, bool) {
	for _, session := range a.sessions {
		if bytes.Compare(token, session.Token) == 0 {
			return session, true
		}
	}

	return AdminSession{}, false
}

//Creates a session with a cryptographically random token
func newAdminSession(username string) AdminSession {
	session := AdminSession{
		Username: username,
		Token:    make([]byte, 256),
	}

	count, err := rand.Read(session.Token)
	if err != nil {
		log.Fatalf("Error generating session token: %s\n", err)
	}
	if count < 256 {
		log.Fatalf("Error generating full length session token\n")
	}

	return session
}

func initHTTP(addr string, port int, certFile string, keyFile string) {
	sessions = AdminSessions{}
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	listenAddr := fmt.Sprintf("%s:%d", addr, port)
	log.Printf("Starting HTTPS server at %s", listenAddr)
	http.ListenAndServeTLS(listenAddr, certFile, keyFile, r)
}

//Authorized middleware to ensure API requests have the session token included
func Authorized(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("API-Token")
		if token == "" {
			//TODO error correctly
			//TODO set proper HTTP status code
			log.Fatalf("Unauthorized")
		} else {
			session, ok := sessions.Get([]byte(token))
			if ok {
				//TODO apply session to context?
				fmt.Printf("sessions %v\n", session)
				next.ServeHTTP(w, r)
			}
		}
	}
	return http.HandlerFunc(fn)
}

//LoginHandler is the HTTP handler that handles authentication to the API
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	type LoginRequestUser struct {
		Username string
		Password string
	}

	var loginRequest LoginRequestUser
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&loginRequest)

	if err != nil {
		//TODO
		log.Fatalf("Error deocidng login request %s\n", err)
	}

	//Read admin information from a file
	var adminUser User
	f, err := os.Open("admin.json")

	if err != nil {
		log.Fatalf("Error reading admin.json %s\n", err)
	}
	adminDec := json.NewDecoder(f)
	adminDec.Decode(&adminUser)

	//Compare username and password to hash
	if adminUser.Username != loginRequest.Username {
		log.Fatalf("Unknown admin user\n")
	}

	err = bcrypt.CompareHashAndPassword(adminUser.Password, []byte(loginRequest.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		//TODO
		log.Fatalf("Password missmatch\n")
	} else if err != nil {
		log.Fatalf("Error comaparing password hash %s\n", err)
	}

	//Create session on success
	session := newAdminSession(loginRequest.Username)
	enc := json.NewEncoder(w)
	err = enc.Encode(session)
	if err != nil {
		log.Fatalf("Error encoding json response %s\n", err)
	}
}
