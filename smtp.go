package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/emersion/go-smtp"
)

//SMTPBackend impliments the SMTP server callback
type SMTPBackend struct{}

//Login for the SMTP backend
func (bkd *SMTPBackend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	fmt.Println("New login")
	if username != "username" || password != "password" {
		return nil, errors.New("Invalid username or password")
	}
	return &Session{}, nil
}

//AnonymousLogin requires clients to authenticate using SMTP AUTH before sending emails
func (bkd *SMTPBackend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	fmt.Println("Anon login")
	return &Session{}, nil
}

//A Session is returned after successful login.
type Session struct{}

//Mail - Initializes a new peice of mail setting the from address
func (s *Session) Mail(from string, opts smtp.MailOptions) error {
	log.Println("Mail from:", from)
	return nil
}

//Rcpt - Add a recipiant to the current message
func (s *Session) Rcpt(to string) error {
	log.Println("Rcpt to:", to)
	return nil
}

//Data - Receive current message data and submit it
func (s *Session) Data(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	log.Println("Data:", string(b))
	return nil
}

//Reset - Discards current message
func (s *Session) Reset() {}

//Logout - Frees all resources for the current session
func (s *Session) Logout() error {
	return nil
}
