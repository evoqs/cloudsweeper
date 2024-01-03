package mail

import (
	"cloudsweep/config"
	logging "cloudsweep/logging"
	notify_model "cloudsweep/notifications/model"
)

type Sender interface {
	Send(from string, to []string, subject, body string, isHTML bool) error
}

type EmailManager struct {
	sender Sender
}

func NewEmailManager(sender Sender) *EmailManager {
	return &EmailManager{sender: sender}
}

func NewDefaultEmailManager() *EmailManager {
	return &EmailManager{sender: NewGomailSender(config.GetConfig().Notifications.Email.Host, config.GetConfig().Notifications.Email.Port,
		config.GetConfig().Notifications.Email.Username, config.GetConfig().Notifications.Email.Password)}
}

func (em *EmailManager) SendEmail(from string, to []string, subject, body string, isHTML bool) error {
	return em.sender.Send(from, to, subject, body, isHTML)
}

func (em *EmailManager) SendNotification(details notify_model.NotfifyDetails) error {
	logging.NewDefaultLogger().Debugf("Processing the Notification from the channel")
	// TODO: Build the body dynamically from the resource
	err := em.SendEmail(config.GetConfig().Notifications.Email.FromAddress, details.EmailDetails.ToAddresses, "Cloud Sweeper Resource Usage Notification", "<html><body><p>Cloud Sweeper Sample Test Mail. You should see the resource details here in future.</p></body></html>", true)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error: %v", err)
	}
	return err
}
