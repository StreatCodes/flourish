package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

//User contains login credentials, alias's, etc.
type User struct {
	Username string
	Password []byte //bcrypted
	Alias    []string
}

//CreateUser is the HTTP handle to add a user to a domain
func CreateUser(w http.ResponseWriter, r *http.Request) {
	domainName := chi.URLParam(r, "domain")

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	type UserInput struct {
		Username string
		Password string
	}
	var user UserInput
	err := dec.Decode(&user)
	if err != nil {
		http.Error(w, "Error deconding JSON body", http.StatusBadRequest)
		return
	}

	domain, err := DomainByName(domainName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = domain.CreateUser(user.Username, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

//ListUsers is the HTTP Handler to list multiple Users
func ListUsers(w http.ResponseWriter, r *http.Request) {
	domainName := chi.URLParam(r, "domain")

	domain, err := DomainByName(domainName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	users := domain.Users()
	enc := json.NewEncoder(w)
	err = enc.Encode(users)
	if err != nil {
		log.Printf("Error encoding json response %s", err)
	}
}

//DeleteUser is the HTTP handle to delete a user from a domain
func DeleteUser(w http.ResponseWriter, r *http.Request) {

}
