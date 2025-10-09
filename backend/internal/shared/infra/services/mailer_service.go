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
	err := smtp.SendMail(fmt.Sprintf("%s:%s", s.host, s.port), s.auth, s.from, m.To, ToBytes(m))
	if err != nil {
		return err
	}
	return nil
}

func ToBytes(m *application.Message) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("Subject: %s\n", m.Subject))
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(m.To, ",")))

	buf.WriteString("MIME-Version: 1.0\n")
	buf.WriteString("Content-Type: text/plain; charset=utf-8\n")

	buf.WriteString(m.Body)

	return buf.Bytes()
}
