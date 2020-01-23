package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
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

//Create adds a session to the active list and returns it
func (a *AdminSessions) Create(username string) AdminSession {
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

	//TODO thread unsafe???
	a.sessions = append(a.sessions, session)
	return session
}

//Initialize the HTTP server
func initHTTP(addr string, port int, certFile string, keyFile string) {
	sessions = AdminSessions{}
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Flourish"))
	})
	r.Post("/login", LoginHandler)

	r.With(Authorized).Post("/domain", CreateDomain)
	r.With(Authorized).Get("/domain", ListDomains)
	r.With(Authorized).Delete("/domain/{domain}", DeleteDomain)

	r.With(Authorized).Post("/domain/{domain}/user", CreateUser)
	r.With(Authorized).Get("/domain/{domain}/user", ListUsers)
	r.With(Authorized).Delete("/domain/{domain}/user/{user}", DeleteUser)

	listenAddr := fmt.Sprintf("%s:%d", addr, port)
	log.Printf("Starting HTTPS server at %s", listenAddr)
	http.ListenAndServeTLS(listenAddr, certFile, keyFile, r)
}

//Authorized middleware to ensure API requests have the session token included
func Authorized(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("API-Token")
		if token == "" {
			log.Printf("API Key required\n")
			http.Error(w, "API Key required", http.StatusUnauthorized)
			return
		}

		decodedToken, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			log.Printf("Error base64 deocding API-Key Header\n")
			http.Error(w, "Error base64 deocding API-Key Header", http.StatusBadRequest)
			return
		}

		session, ok := sessions.Get(decodedToken)
		if ok {
			//TODO apply session to context?
			next.ServeHTTP(w, r)
		} else {
			log.Printf("Invalid API key: %s\n", string(session.Token))
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
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
	defer r.Body.Close()

	err := dec.Decode(&loginRequest)
	if err != nil {
		log.Printf("Error decoding login request %s\n", err)
		http.Error(w, "Error decoding json login request", http.StatusBadRequest)
		return
	}

	//Read admin information from a file
	var adminUser User
	f, err := os.Open("admin.json")

	if err != nil {
		log.Printf("Error reading admin.json %s\n", err)
		http.Error(w, "Error reading admin users", http.StatusInternalServerError)
		return
	}
	adminDec := json.NewDecoder(f)
	adminDec.Decode(&adminUser)

	//Compare username and password to hash
	if adminUser.Username != loginRequest.Username {
		log.Printf("Unknown admin user: %s\n", loginRequest.Username)
		http.Error(w, "Unknown user", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword(adminUser.Password, []byte(loginRequest.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		http.Error(w, "Password doesn't match.", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Error comparing password hash %s\n", err)
		http.Error(w, "Unexpected bcrypt error", http.StatusInternalServerError)
		return
	}

	//Create session on success
	session := sessions.Create(loginRequest.Username)
	enc := json.NewEncoder(w)
	err = enc.Encode(session)
	if err != nil {
		log.Printf("Error encoding json response %s\n", err)
		http.Error(w, "Unexpected json encoding error", http.StatusInternalServerError)
		return
	}
}
