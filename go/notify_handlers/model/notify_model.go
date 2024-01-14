package notifications

// Structure of the Resource detail for Notification
type NotifyResourceDetails struct {
	AccountID               string            `json:"accountId"`
	ResourceClass           string            `json:"resourceClass"`
	ResourceId              string            `json:"resourceId"`
	ResourceName            string            `json:"resourceName"`
	CurrentResourceType     string            `json:"currentResourceType"`
	RecommendedResourceType string            `json:"recommendedResourceType"`
	RegionCode              string            `json:"regionCode"`
	MonthlyPrice            float64           `json:"monthlyPrice"`
	ResourceTags            map[string]string `json:"resourceTags"`
	MonthlySavings          float64           `json:"monthlySavings"`
}

type NotifyRequest struct {
	Channels []string
	Details  NotfifyDetails
}

type NotfifyDetails struct {
	PipeLineName    string                  `json:"pipeLineName"`
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
