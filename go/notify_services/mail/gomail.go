package mail

import (
	logging "cloudsweep/logging"
	"io"

	"gopkg.in/gomail.v2"
)

type GomailSender struct {
	dialer *gomail.Dialer
}

func NewGomailSender(host string, port int, username, password string) *GomailSender {
	return &GomailSender{
		dialer: gomail.NewDialer(host, port, username, password),
	}
}

func (gs *GomailSender) Send(emailDetails EmailDetails) error {
	logging.NewDefaultLogger().Debugf("Sending the mail to %v from %s subject %s", emailDetails.To, emailDetails.From, emailDetails.Subject) //, emailDetails.BodyHTML)
	m := gomail.NewMessage()
	m.SetHeader("From", emailDetails.From)
	m.SetHeader("To", emailDetails.To...)
	m.SetHeader("Subject", emailDetails.Subject)

	if emailDetails.BodyHTML != "" {
		m.SetBody("text/html", emailDetails.BodyHTML)
	} else {
		m.SetBody("text/plain", emailDetails.BodyPlain)
	}

	if emailDetails.AttachmentName != "" && len(emailDetails.AttachmentData) > 0 {
		m.Attach(emailDetails.AttachmentName, gomail.SetCopyFunc(func(writer io.Writer) error {
			_, err := writer.Write(emailDetails.AttachmentData)
			return err
		}))
	}

	// Embed images
	for _, imageLocation := range emailDetails.ImageLocations {
		m.Embed(imageLocation)
	}

	err := gs.dialer.DialAndSend(m)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error sending email: %v", err)
		return err
	}
	return nil
}
