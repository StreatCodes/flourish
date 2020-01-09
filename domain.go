package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path"
	"sync"
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
					log.Fatalf("Error reading flourish.data for %s: %e\n", f.Name(), err)
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
		log.Fatal(err)
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
		log.Fatalf("Couldn't encode JSON response %e", err)
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
		//TODO probably want to put erros in context and handle with middleware
		log.Fatalf("Received invalid JSON: %e\n", err)
	}

	//TODO validate domain name?
	_, err = mail.ParseAddress("john.doe@" + domain.Name)
	if err != nil {
		//TODO
		log.Fatalf("Invalid domain: %e", err)
	}

	//Create domain directory
	dPath := path.Join(config.MailDir, domain.Name)
	err = os.Mkdir(dPath, 644)
	if err == os.ErrExist {
		//TODO handle error correctly
		log.Fatalf("Domain already exists")
	} else if err != nil {
		//TODO handle error correctly
		log.Fatalf("Unexpected error when creating domain: %e", err)
	}

	w.WriteHeader(200)
}
