package mail

type EmailDetails struct {
	To             []string
	From           string
	Subject        string
	BodyPlain      string
	BodyHTML       string
	AttachmentName string
	AttachmentData []byte
	ImageLocations []string
}
