package mail

import (
	"bytes"
	logging "cloudsweep/logging"
	"fmt"
	"log"
	"mime/multipart"
	"net/textproto"

	"gopkg.in/mail.v2"
)

type GomailSender struct {
	dialer *mail.Dialer
}

func NewGomailSender(host string, port int, username, password string) *GomailSender {
	return &GomailSender{
		dialer: mail.NewDialer(host, port, username, password),
	}
}
func (gs *GomailSender) SendWithAttachment(from string, to []string, subject, bodyHTML string, attachmentName string, attachmentData []byte) error {
	m := mail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)

	// Create a multipart message
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add HTML part
	htmlPart := make(textproto.MIMEHeader)
	htmlPart.Set("Content-Type", "text/html")
	htmlPart.Set("Content-Transfer-Encoding", "quoted-printable")
	htmlPartWriter, _ := writer.CreatePart(htmlPart)
	htmlPartWriter.Write([]byte(bodyHTML))

	// Add Attachment part
	attachmentPart := make(textproto.MIMEHeader)
	attachmentPart.Set("Content-Type", "application/octet-stream")
	attachmentPart.Set("Content-Transfer-Encoding", "base64")
	attachmentPart.Set("Content-Disposition", `attachment; filename="`+attachmentName+`"`)
	attachmentPartWriter, _ := writer.CreatePart(attachmentPart)
	attachmentPartWriter.Write(attachmentData)

	// Close the multipart writer to finalize the message
	writer.Close()

	m.SetBody("multipart/mixed", body.String())

	fmt.Printf("Sending mail.....")
	err := gs.dialer.DialAndSend(m)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error sending email..: %s", err)
		return err
	}
	return nil
}

func (gs *GomailSender) Send(from string, to []string, subject, body string, isHTML bool) error {
	logging.NewDefaultLogger().Debugf("Sending the mail to %v from %s subject %s body %s", to, from, subject, body)
	m := mail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	if isHTML {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	err := gs.dialer.DialAndSend(m)
	if err != nil {
		log.Println("Error sending email:", err)
		return err
	}
	return nil
}
