package mail

import (
	"bytes"
	"cloudsweep/config"
	logging "cloudsweep/logging"
	notify_model "cloudsweep/notifications/model"
	"html/template"
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
	body, err := buildEmailBody(details)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Problem in building the email body %v", err)
	}
	err = em.SendEmail(config.GetConfig().Notifications.Email.FromAddress, details.EmailDetails.ToAddresses, "Cloud Sweeper Resource Usage Notification", body, true)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error: %v", err)
	}
	return err
}

func buildEmailBody(details notify_model.NotfifyDetails) (string, error) {
	// Load your company logo
	logoURL := "/home/pavan/Documents/cs.jpg"
	// Calculate the sum of Monthly Prices
	var totalMonthlyPrice float64
	var totalMonthlySavings float64
	for _, resource := range details.ResourceDetails {
		totalMonthlyPrice += resource.MonthlyPrice
		totalMonthlySavings += resource.MonthlySavings
	}
	// Template for the email body
	emailTemplate := `
	<html>
		<head>
			<style>
				.logo {
					max-width: 200px;
					height: auto;
					display: block;
					margin: 20px auto;
				}

				.button {
					display: inline-block;
					font-size: 16px;
					padding: 10px 20px;
					text-align: center;
					text-decoration: none;
					background-color: #3b6696;
					color: white;
					border-radius: 5px;
				}

				/* External table styling */
				.external-table {
					border-collapse: collapse;
					width: 100%;
					border: 0.5px solid #d3d3d3; /* External border color */
					border-radius: 20px; /* Rounded corners */
				}

				.external-table td, .external-table th {
					padding: 10px;
					background-color: #f8f8f8; /* Internal color */
				}

				/* External table heading */
					.external-table-heading {
					text-align: center;
					color: #2f2f2f;
				}

				/* Internal tables styling */
				.internal-table {
					border-collapse: collapse;
					width: 100%;
					margin-top: 20px;
					border: 0.5px solid #dddddd;
				}

				.internal-table th, .internal-table td {
					text-align: left;
					padding: 8px;
					background-color: #f8f8f8; /* Internal color */
					border: none;
				}

				.internal-table th {
					background-color: #e8e8e8;
				}
			</style>
		</head>

		<body>
			<tr>
					<td colspan="2" style="text-align: center;">
						<img src="{{.LogoURL}}" alt="Company Logo" class="logo">
					</td>
				</tr>
			<table class="external-table">
				<tr class="external-table-heading">
					<td colspan="2" style="text-align: center;">
						<h1>Resources Summary</h1>
					</td>
				</tr>
				<tr>
					<td style="text-align: left;">
						<span>Pipeline Name:</span>
						<h3 style="margin: 0;">{{.PipeLineName}}</h3>
					</td>
					<td style="text-align: right;">
						<span>Monthly Price:</span>
						<h3 style="margin: 0;">{{printf "$%.2f" .TotalMonthlyPrice}}</h3>
					</td>
				</tr>
				<tr>
					<td style="text-align: right;"> </td>
					<td style="text-align: right;">
						<span>Monthly Savings:</span>
						<h3 style="margin: 0;">{{printf "$%.2f" .TotalMonthlySavings}}</h3>
					</td>
				</tr>
				<tr>
					<td colspan="2">
						<table class="internal-table">
							<tr>
								<th>Account ID</th>
								<th>Resource Type</th>
								<th>Resource ID</th>
								<th>Resource Name</th>
								<th>Region Code</th>
								<th>Monthly Price</th>
								<th>Recommendation</th>
								<th>Monthly Savings</th>
							</tr>
							{{range .ResourceDetails}}
							<tr>
								<td>{{.AccountID}}</td>
								<td>{{.ResourceType}}</td>
								<td>{{.ResourceId}}</td>
								<td>{{.ResourceName}}</td>
								<td>{{.RegionCode}}</td>
								<td>{{.MonthlyPrice}}</td>
								<td>{{.Recommendation}}</td>
								<td>{{.MonthlySavings}}</td>
							</tr>
							{{end}}
						</table>
					</td>
				</tr>
			</table>

			<div style="text-align: center; margin-top: 20px;">
				<a href="{{.CompanyURL}}" class="button" style="background-color: #3b6696; color: white;" target="_blank">Visit CloudSweeper</a>
			</div>
		</body>
	</html>`

	// Prepare data for the template
	data := struct {
		LogoURL             string
		CompanyURL          string
		ResourceDetails     []notify_model.NotifyResourceDetails
		PipeLineName        string
		TotalMonthlyPrice   float64
		TotalMonthlySavings float64
	}{
		LogoURL:             logoURL,
		CompanyURL:          "https://cloudsweeper.in/",
		ResourceDetails:     details.ResourceDetails,
		PipeLineName:        details.PipeLineName,
		TotalMonthlyPrice:   totalMonthlyPrice,
		TotalMonthlySavings: totalMonthlySavings,
	}

	// Execute the template
	tmpl, err := template.New("emailTemplate").Parse(emailTemplate)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error parsing email template: %v", err)
		return "", err
	}

	var emailBodyBuffer bytes.Buffer
	err = tmpl.Execute(&emailBodyBuffer, data)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error executing email template: %v", err)
		return "", err
	}

	return emailBodyBuffer.String(), nil
}
