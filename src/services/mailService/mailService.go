package mailService

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"gopkg.in/gomail.v2"
)

func SendMail(email string, subject string, body string) {
	if os.Getenv("GO_ENV") == "test" {
		return
	}

	user := os.Getenv("MAILING_EMAIL")
	pass := os.Getenv("MAILING_PASSWORD")

	m := gomail.NewMessage()
	m.SetHeader("From", user)
	m.SetHeader("To", email)
	m.SetHeader("Subject", fmt.Sprintf("i9chat - %s", subject))
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 465, user, pass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: os.Getenv("GO_ENV") != "production"}

	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
		return
	}
}
