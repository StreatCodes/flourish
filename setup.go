package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

//Setup gathers user input for initial server setup
func Setup() {
	//Get email and password from user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the administrator email: ")
	adminEmail, _ := reader.ReadString('\n')
	fmt.Print("Enter the administrator password: ")
	adminPassword, _ := reader.ReadString('\n')

	//Hash password
	pw := []byte(adminPassword[:len(adminPassword)-1])
	encryptedPassword, err := bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)

	if err != nil {
		log.Fatalf("Error generating password hash: %s", err)
	}

	adminUser := User{
		Username: adminEmail[:len(adminEmail)-1],
		Password: encryptedPassword,
	}

	//Open on create admin.json
	adminFileName := "admin.json"
	f, err := os.OpenFile(adminFileName, os.O_RDWR, 644)
	if os.IsNotExist(err) {
		f, err = os.Create(adminFileName)
		if err != nil {
			log.Fatalf("Failed to create %s: %s", adminFileName, err)
		}
	} else if err != nil {
		log.Fatalf("Failed to open %s: %s", adminFileName, err)
	}

	//Encode credentials to file
	enc := json.NewEncoder(f)
	err = enc.Encode(adminUser)
	if err != nil {
		log.Fatalf("Failed to open %s: %s", adminFileName, err)
	}

	fmt.Println("Admin config generated")
}
