package mail

type Sender interface {
	Send(from string, to []string, subject, body string, isHTML bool) error
}

type Manager struct {
	sender Sender
}

func NewEmailManager(sender Sender) *Manager {
	return &Manager{sender: sender}
}

func (em *Manager) SendEmail(from string, to []string, subject, body string, isHTML bool) error {
	return em.sender.Send(from, to, subject, body, isHTML)
}
