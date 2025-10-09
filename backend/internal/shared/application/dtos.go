package application

type Message struct {
	To      []string
	Subject string
	Body    string
}

type MailerConfig struct {
	SMTP_HOST     string
	SMTP_PORT     string
	SMTP_FROM     string
	SMTP_PASSWORD string
}
