package mailer

import (
	"bytes"
	"fmt"
	"net/smtp"
)

type M struct {
	serverAddr    string
	defaultSender string
}

func New(serverAddr string, defaultSender string) *M {
	return &M{
		serverAddr,
		defaultSender,
	}
}

func (m M) SendMail(recipient string, subject string, body string) error {
	return Send(m.serverAddr, m.defaultSender, recipient, subject, body)
}

func Send(serverAddr, from, to, subject, body string) error {
	c, err := smtp.Dial(serverAddr)
	defer c.Close()
	if err != nil {
		return err
	}

	err = c.Mail(from)
	if err != nil {
		return err
	}

	err = c.Rcpt(to)
	if err != nil {
		return err
	}

	wc, err := c.Data()
	defer wc.Close()
	if err != nil {
		return err
	}

	buf := bytes.NewBufferString(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))
	if _, err = buf.WriteTo(wc); err != nil {
		return err
	}

	return nil
}
