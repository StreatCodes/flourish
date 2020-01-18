package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path"
	"sync"

	"github.com/go-chi/chi"
)

//LoadDomains builds a Domains struct from the mailDir directory
func LoadDomains(mailDir string) *Domains {
	//Load domains
	files, err := ioutil.ReadDir(mailDir)
	if err != nil {
		log.Fatal(err)
	}

	//Look for all domains and register them
	domains := Domains{
		domains: make(map[string]Domain),
	}

	for _, f := range files {
		if f.IsDir() {
			dbPath := path.Join(config.MailDir, f.Name(), "flourish.data")
			if file, err := os.Open(dbPath); err == nil {
				fmt.Printf("Found domain %s\n", f.Name())

				//Decode flourish domain data
				var domain Domain
				dec := json.NewDecoder(file)
				err := dec.Decode(&domain)
				if err != nil {
					log.Fatalf("Error reading flourish.data for %s: %s\n", f.Name(), err)
				}
				domain.Name = f.Name()
				domains.add(f.Name(), domain)
			}
		}
	}

	return &domains
}

//Domains contains all the available domains for the Mail Server
type Domains struct {
	mutex   sync.Mutex
	domains map[string]Domain
}

func (d *Domains) get(domainName string) Domain {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.domains[domainName]
}

func (d *Domains) add(domainName string, domain Domain) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.domains[domainName] = domain
}

func (d *Domains) del(domainName string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.domains, domainName)
}

//Domain contains information about it's users
type Domain struct {
	Name  string
	Users []User
}

//ListDomains is the HTTP Handler to list multiple Domains
func ListDomains(w http.ResponseWriter, r *http.Request) {
	//Load domains
	files, err := ioutil.ReadDir(config.MailDir)
	if err != nil {
		log.Printf("Error reading domain directory structure %s\n", err)
		http.Error(w, "Error reading domain directory structure", http.StatusInternalServerError)
		return
	}

	var domains []string

	for _, f := range files {
		if f.IsDir() {
			domains = append(domains, f.Name())
		}
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(domains)
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
	err = os.Mkdir(dPath, 644)
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
	domain := chi.URLParam(r, "domain")

	//Delete domain and all sub-directories
	dPath := path.Join(config.MailDir, domain)

	//Check domain exists
	_, err := os.Stat(dPath)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("Domain does not exist\n")
		http.Error(w, "Domain does not exist", http.StatusBadRequest)
		return
	}

	//Delete domain
	err = os.RemoveAll(dPath)
	if err != nil {
		log.Printf("Unexpected error when deleting domain: %s\n", err)
		http.Error(w, "Unexpected error when deleting domain", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
