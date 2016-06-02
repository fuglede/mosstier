package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/smtp"
)

// sendMail sends a mail to a given recipient with given contents. Details on
// SMTP connection and sender information must be specified in the config.
// Note: To send mails to multiple recipients, simply use this function several
// times; if the number of users grows too large, it might be useful to go to a
// lower level and keep the connection alive.
func sendMail(recipient string, subject string, body string) (err error) {
	// Part of this comes from https://gist.github.com/andelf/5004821
	header := make(map[string]string)
	header["To"] = recipient
	header["From"] = config.MailSender
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	auth := smtp.PlainAuth(
		"",
		config.SMTPUsername,
		config.SMTPPassword,
		config.SMTPHost,
	)

	err = smtp.SendMail(
		fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort),
		auth,
		config.MailSender,
		[]string{recipient},
		[]byte(message),
	)
	if err != nil {
		log.Println("Could not send mail: ", err)
	}
	return
}

// sendMails sends mails to several recipients
func sendMails(recipients []string, subject string, body string) error {
	var err error
	for _, recipient := range recipients {
		err = sendMail(recipient, subject, body)
		if err != nil {
			log.Println("Could send mail to "+recipient+": ", err)
		}
	}
	if err != nil {
		return errors.New("could not send to one or more recipients")
	}
	return nil
}
