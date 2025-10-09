package application

type MailerService interface {
	Send(m *Message, config MailerConfig) error
}
