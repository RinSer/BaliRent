package mail

import (
	"net/smtp"
	"os"
)

type Mail struct {
	From string
	Password string
	Smtp string
}

func NewMail() *Mail {
	return &Mail{
		From: os.Getenv("MAIL_FROM"),
		Password: os.Getenv("MAIL_PASSWORD"),
		Smtp: os.Getenv("MAIL_SERVER"),
	}
}

func (o *Mail) Send(to string, subject string, body string) error {
    var auth smtp.Auth = nil
	if o.Password != "" {
		auth = smtp.PlainAuth("", o.From, o.Password,o.Smtp)
	}
	
    err :=  smtp.SendMail(o.Smtp, auth, o.From, []string{to}, []byte(body))
	return err
  
}