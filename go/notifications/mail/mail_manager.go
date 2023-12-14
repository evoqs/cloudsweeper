package mail

type Sender interface {
	Send(to []string, subject, body string) error
}

type Manager struct {
	sender Sender
}

func NewEmailManager(sender Sender) *Manager {
	return &Manager{sender: sender}
}

func (em *Manager) SendEmail(to []string, subject, body string) error {
	return em.sender.Send(to, subject, body)
}
