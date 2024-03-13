package appservices

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func SendMail(email string, subject string, body string) error {
	to := []string{email}
	from := os.Getenv("MAILING_EMAIL")

	auth := smtp.PlainAuth("", from, os.Getenv("MAILING_PASSWORD"), "smtp.gmail.com")

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: i9chat - %s\r\n\r\n%s\r\n", email, subject, body))

	err := smtp.SendMail("smtp.gmail.com:465", auth, from, to, msg)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
