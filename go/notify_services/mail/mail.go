package mail

import (
	"bytes"
	logging "cloudsweep/logging"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"

	"gopkg.in/mail.v2"
)

type MailSender struct {
	dialer *mail.Dialer
}

func NewMailSender(host string, port int, username, password string) *MailSender {
	return &MailSender{
		dialer: mail.NewDialer(host, port, username, password),
	}
}

// TODO: Convert the parameters to emailDetails Model
func (gs *MailSender) SendWithAttachment(from string, to []string, subject, bodyHTML string, attachmentName string, attachmentData []byte) error {
	logoURL := "/home/pavan/Documents/cs.jpg"
	m := mail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.Embed(logoURL)

	// Create a multipart message
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add HTML part
	htmlPart := make(textproto.MIMEHeader)
	htmlPart.Set("Content-Type", "text/html")
	htmlPart.Set("Content-Transfer-Encoding", "quoted-printable")
	htmlPartWriter, err := writer.CreatePart(htmlPart)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error creating html part: %s", err)
		return err
	}
	htmlPartWriter.Write([]byte(bodyHTML))

	// Add Attachment part
	attachmentPart := make(textproto.MIMEHeader)
	attachmentPart.Set("Content-Type", "application/octet-stream")
	attachmentPart.Set("Content-Transfer-Encoding", "base64")
	attachmentPart.Set("Content-Disposition", `attachment; filename="`+attachmentName+`"`)
	attachmentPartWriter, err := writer.CreatePart(attachmentPart)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error creating attachment part: %s", err)
		return err
	}
	attachmentPartWriter.Write(attachmentData)

	// Close the multipart writer to finalize the message
	writer.Close()

	m.SetBody("multipart/mixed", body.String())

	fmt.Printf("Sending mail.....")
	err = gs.dialer.DialAndSend(m)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error sending email..: %s", err)
		return err
	}
	return nil
}

func (gs *MailSender) Send(emailDetails EmailDetails) error {
	logging.NewDefaultLogger().Debugf("Sending the mail to %v from %s subject %s body %s", emailDetails.To, emailDetails.From, emailDetails.Subject, emailDetails.BodyHTML)
	m := mail.NewMessage()
	m.SetHeader("From", emailDetails.From)
	m.SetHeader("To", emailDetails.To...)
	m.SetHeader("Subject", emailDetails.Subject)

	if emailDetails.BodyHTML != "" {
		m.SetBody("text/html", emailDetails.BodyHTML)
	} else {
		m.SetBody("text/plain", emailDetails.BodyPlain)
	}

	if emailDetails.AttachmentName != "" && len(emailDetails.AttachmentData) > 0 {
		m.Attach(emailDetails.AttachmentName, mail.SetCopyFunc(func(writer io.Writer) error {
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
