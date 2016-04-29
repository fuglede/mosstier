package main

import (
	"encoding/base64"
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
		config.SmtpUsername,
		config.SmtpPassword,
		config.SmtpHost,
	)

	err = smtp.SendMail(
		fmt.Sprintf("%s:%d", config.SmtpHost, config.SmtpPort),
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
