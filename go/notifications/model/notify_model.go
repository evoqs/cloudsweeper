package model

// Structure of the Resource detail for Notification
type NotifyResourceDetails struct {
	AccountID    string            `json:"accountId"`
	ResourceType string            `json:"resourceType"`
	ResourceId   string            `json:"resourceId"`
	ResourceName string            `json:"resourceName"`
	RegionCode   string            `json:"regionCode"`
	ResourceTags map[string]string `json:"resourceTags"`
}

type NotifyRequest struct {
	Channels []string
	Details  NotfifyDetails
}

type NotfifyDetails struct {
	ResourceDetails []NotifyResourceDetails `json:"resourceDetails"`
	EmailDetails    NotifyEmailDetails      `json:"emailDetails"`
	SlackDetails    NotifySlackDetails      `json:"slackDetails"`
	WebhookDetails  NotifyWebhookDetails    `json:"webhookDetails"`
}

type NotifyEmailDetails struct {
	Enabled     bool     `json:"enabled"`
	ToAddresses []string `json:"toAddresses"`
}

type NotifySlackDetails struct {
	Enabled bool     `json:"enabled"`
	Url     []string `json:"url"`
}

type NotifyWebhookDetails struct {
	Enabled bool     `json:"enabled"`
	Url     []string `json:"url"`
}
