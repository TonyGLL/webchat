package application

type MailerService interface {
	Send(m *Message) error
}
