package application

type Message struct {
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	Attachments map[string][]byte
}

type MailerConfig struct {
	SMTP_HOST     string
	SMTP_PORT     string
	SMTP_FROM     string
	SMTP_PASSWORD string
}
