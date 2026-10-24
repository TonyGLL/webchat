package services

import (
	"backend/internal/shared/application"
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

const (
	// smtpDialTimeout bounds the TCP connect; smtpDeadline bounds the whole conversation.
	// smtp.SendMail has neither, so a black-holed SMTP host would hang the caller forever.
	smtpDialTimeout = 10 * time.Second
	smtpDeadline    = 30 * time.Second
)

type MailerService struct {
	auth smtp.Auth
	host string
	port string
	from string

	dialTimeout time.Duration
	deadline    time.Duration
}

func NewMailerService(config application.MailerConfig) (application.MailerService, error) {
	auth := smtp.PlainAuth("", config.SMTP_FROM, config.SMTP_PASSWORD, config.SMTP_HOST)
	return &MailerService{
		auth: auth,
		host: config.SMTP_HOST,
		port: config.SMTP_PORT,
		from: config.SMTP_FROM,

		dialTimeout: smtpDialTimeout,
		deadline:    smtpDeadline,
	}, nil
}

func (s *MailerService) Send(m *application.Message) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	msg := ToBytes(s.from, m)
	return s.send(addr, msg, m.To)
}

// send mirrors smtp.SendMail (STARTTLS when offered, then AUTH) but dials with a timeout and
// puts a deadline on the whole conversation.
func (s *MailerService) send(addr string, msg []byte, to []string) error {
	conn, err := (&net.Dialer{Timeout: s.dialTimeout}).Dial("tcp", addr)
	if err != nil {
		return err
	}
	if err := conn.SetDeadline(time.Now().Add(s.deadline)); err != nil {
		conn.Close()
		return err
	}
	c, err := smtp.NewClient(conn, s.host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()

	if err := c.Hello("localhost"); err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return err
		}
	}
	if s.auth != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return fmt.Errorf("smtp: server doesn't support AUTH")
		}
		if err := c.Auth(s.auth); err != nil {
			return err
		}
	}
	if err := c.Mail(s.from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func ToBytes(from string, m *application.Message) []byte {
	var buf bytes.Buffer
	crlf := "\r\n"

	// Headers (order matters)
	buf.WriteString(fmt.Sprintf("From: %s%s", from, crlf))
	buf.WriteString(fmt.Sprintf("To: %s%s", strings.Join(m.To, ","), crlf))
	buf.WriteString(fmt.Sprintf("Subject: %s%s", m.Subject, crlf))
	buf.WriteString("MIME-Version: 1.0" + crlf)
	buf.WriteString("Content-Type: text/plain; charset=utf-8" + crlf)
	buf.WriteString("Content-Transfer-Encoding: 7bit" + crlf)

	// Blank line between headers and body
	buf.WriteString(crlf)

	// Message body
	buf.WriteString(m.Body)
	buf.WriteString(crlf)

	return buf.Bytes()
}
