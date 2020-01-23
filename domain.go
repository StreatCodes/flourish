package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path"

	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
)

//Domain contains information about it's users
type Domain struct {
	Name string
}

//Users gets all users associated with a domain
func (d *Domain) Users() []User {
	dPath := path.Join(config.MailDir, d.Name)
	//Load users
	files, err := ioutil.ReadDir(dPath)
	if err != nil {
		log.Fatal(err) //TODO
	}

	//Look for all users
	var users []User

	for _, f := range files {
		if f.IsDir() {
			dbPath := path.Join(dPath, f.Name(), "flourish-user.json")
			if file, err := os.Open(dbPath); err == nil {
				//Decode flourish user data
				var user User
				dec := json.NewDecoder(file)
				err := dec.Decode(&user)
				if err != nil {
					log.Fatalf("Error reading flourish-user.json for %s@%s: %s\n", f.Name(), d.Name, err)
				}
				user.Username = f.Name()
				user.Password = nil

				users = append(users, user)
			}
		}
	}

	return users
}

//CreateUser gets all users associated with a domain
func (d *Domain) CreateUser(username string, password string) (User, error) {
	//Validate user + domain
	_, err := mail.ParseAddress(username + "@" + d.Name)
	if err != nil {
		log.Printf("Invalid domain: %s\n", err)
		return User{}, errors.New("Invalid username")
	}

	//Create user directory
	uPath := path.Join(config.MailDir, d.Name, username)
	err = os.Mkdir(uPath, 0700)
	if errors.Is(err, os.ErrExist) {
		return User{}, errors.New("User already exists")
	} else if err != nil {
		return User{}, errors.New("Unexpected error creating user: " + err.Error())
	}

	encryptedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.New("Error generating new user's password hash: " + err.Error())
	}

	user := User{
		Username: username,
		Password: encryptedPass,
		Alias:    make([]string, 0),
	}

	flourishDataPath := path.Join(uPath, "flourish-user.json")
	f, err := os.OpenFile(flourishDataPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return User{}, errors.New("Error creating user data file: " + err.Error())
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	err = enc.Encode(user)
	if err != nil {
		return User{}, errors.New("Error writing user data: " + err.Error())
	}

	return user, nil
}

//DeleteUser deletes the user with the given name that's associated with the domain
func (d *Domain) DeleteUser(username string) error {
	uPath := path.Join(config.MailDir, d.Name, username)

	//Delete domain
	err := os.RemoveAll(uPath)
	if err != nil {
		return err
	}

	return nil
}

//AllDomains searchs the MailDir directory for domains
func AllDomains() []Domain {
	//Load domains
	files, err := ioutil.ReadDir(config.MailDir)
	if err != nil {
		log.Fatal(err)
	}

	//Look for all domains
	var domains []Domain

	for _, f := range files {
		if f.IsDir() {
			domains = append(domains, Domain{Name: f.Name()})
		}
	}

	return domains
}

//DeleteDomainByName deletes the domain, users and all user data
func DeleteDomainByName(name string) error {
	domain, err := DomainByName(name)
	if err != nil {
		return err
	}

	for _, user := range domain.Users() {
		err = domain.DeleteUser(user.Username)
		if err != nil {
			return errors.New("Failed to delete user " + user.Username + " " + err.Error())
		}
	}

	uPath := path.Join(config.MailDir, name)

	//Delete domain
	err = os.RemoveAll(uPath)
	if err != nil {
		return err
	}

	return nil
}

//DomainByName returns the domain associated with the given name
func DomainByName(name string) (Domain, error) {
	dPath := path.Join(config.MailDir, name)
	f, err := os.Stat(dPath)
	if err != nil {
		return Domain{}, err
	}

	if !f.IsDir() {
		return Domain{}, errors.New("Domain not found")
	}

	return Domain{Name: f.Name()}, nil
}

//ListDomains is the HTTP Handler to list multiple Domains
func ListDomains(w http.ResponseWriter, r *http.Request) {
	domains := AllDomains()

	enc := json.NewEncoder(w)
	err := enc.Encode(domains)
	if err != nil {
		log.Printf("Error encoding JSON response %s\n", err)
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}
}

//CreateDomain is the HTTP handler to create a Domain
func CreateDomain(w http.ResponseWriter, r *http.Request) {
	type DomainBody struct {
		Name string
	}

	var domain DomainBody
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := dec.Decode(&domain)
	if err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	//Validate domain
	_, err = mail.ParseAddress("john.doe@" + domain.Name)
	if err != nil {
		log.Printf("Invalid domain: %s\n", err)
		http.Error(w, "Invalid domain", http.StatusBadRequest)
		return
	}

	//Create domain directory
	dPath := path.Join(config.MailDir, domain.Name)
	err = os.Mkdir(dPath, 0700)
	if errors.Is(err, os.ErrExist) {
		log.Printf("Domain already exists\n")
		http.Error(w, "Domain already exists", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Printf("Unexpected error when creating domain: %s\n", err)
		http.Error(w, "Unexpected error when creating domain", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

//DeleteDomain is the HTTP handler to delete a Domain
func DeleteDomain(w http.ResponseWriter, r *http.Request) {
	domainName := chi.URLParam(r, "domain")

	//Delete domain
	err := DeleteDomainByName(domainName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "Domain does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
