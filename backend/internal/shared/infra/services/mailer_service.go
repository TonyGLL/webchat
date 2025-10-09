package services

import (
	"backend/internal/shared/application"
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"strings"
)

type MailerService struct {
	auth smtp.Auth
}

func NewMailerService(config application.MailerConfig) (application.MailerService, error) {
	auth := smtp.PlainAuth("", config.SMTP_FROM, config.SMTP_PASSWORD, config.SMTP_HOST)
	return &MailerService{auth}, nil
}

func (s *MailerService) Send(m *application.Message, config application.MailerConfig) error {
	return smtp.SendMail(fmt.Sprintf("%s:%s", config.SMTP_HOST, config.SMTP_PORT), s.auth, config.SMTP_FROM, m.To, ToBytes(m))
}

func ToBytes(m *application.Message) []byte {
	buf := bytes.NewBuffer(nil)
	withAttachments := len(m.Attachments) > 0
	buf.WriteString(fmt.Sprintf("Subject: %s\n", m.Subject))
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(m.To, ",")))
	if len(m.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\n", strings.Join(m.CC, ",")))
	}

	if len(m.BCC) > 0 {
		buf.WriteString(fmt.Sprintf("Bcc: %s\n", strings.Join(m.BCC, ",")))
	}

	buf.WriteString("MIME-Version: 1.0\n")
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	if withAttachments {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", boundary))
		buf.WriteString(fmt.Sprintf("--%s\n", boundary))
	} else {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\n")
	}

	buf.WriteString(m.Body)
	if withAttachments {
		for k, v := range m.Attachments {
			buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\n", http.DetectContentType(v)))
			buf.WriteString("Content-Transfer-Encoding: base64\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\n", k))

			b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(b, v)
			buf.Write(b)
			buf.WriteString(fmt.Sprintf("\n--%s", boundary))
		}

		buf.WriteString("--")
	}

	return buf.Bytes()
}
