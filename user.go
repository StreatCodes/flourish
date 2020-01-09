package main

//User contains login credentials, alias's, etc.
type User struct {
	Username string
	Password []byte //bcrypted
}
