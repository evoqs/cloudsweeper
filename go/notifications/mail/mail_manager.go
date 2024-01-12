package mail

import (
	"cloudsweep/config"
	logging "cloudsweep/logging"
)

type Sender interface {
	//Send(from string, to []string, subject, body string, isHTML bool) error
	Send(emailDetails EmailDetails) error
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

func (em *EmailManager) SendEmail(emailDetails EmailDetails) error {
	return em.sender.Send(emailDetails)
}

func (em *EmailManager) SendNotification(details EmailDetails) error {
	logging.NewDefaultLogger().Debugf("Processing the Notification from the channel")
	// Build the body dynamically from the resource
	/*body, err := buildEmailBody(details)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Problem in building the email body %v", err)
	}
	// Create a CSV file with header and details
	csvData, err := createAttachmentDataInCsvFormat(details.ResourceDetails)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error creating CSV file: %v", err)
		return err
	}
	EmailDetails{
		To:        details.To,
		From:      config.GetConfig().Notifications.Email.FromAddress,
		Subject:   "Cloud Sweeper Resource Usage Notification",
		BodyHTML:  details.BodyHTML,
		BodyPlain: details.BodyPlain,
		// TODO: Get this from config
		ImageLocations: []string{"/home/pavan/Documents/cs.jpg"},
		AttachmentName: "resource_details.csv",
		AttachmentData: []byte(csvData),
	}
	*/
	err := em.SendEmail(details)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error: %v", err)
	}
	return err
}
