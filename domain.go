package main

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
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
				dec := gob.NewDecoder(file)
				dec.Decode(&domain)
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

type Domain struct {
	Name  string
	Users []User
}
