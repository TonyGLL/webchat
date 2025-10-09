package services

import (
	"backend/internal/shared/application"
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
)

type MailerService struct {
	auth smtp.Auth
	host string
	port string
	from string
}

func NewMailerService(config application.MailerConfig) (application.MailerService, error) {
	auth := smtp.PlainAuth("", config.SMTP_FROM, config.SMTP_PASSWORD, config.SMTP_HOST)
	return &MailerService{
		auth: auth,
		host: config.SMTP_HOST,
		port: config.SMTP_PORT,
		from: config.SMTP_FROM,
	}, nil
}

func (s *MailerService) Send(m *application.Message) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	msg := ToBytes(s.from, m)
	return smtp.SendMail(addr, s.auth, s.from, m.To, msg)
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
