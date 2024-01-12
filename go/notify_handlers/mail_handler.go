package notifications

import (
	"bytes"
	"cloudsweep/config"
	logging "cloudsweep/logging"
	model "cloudsweep/notify_handlers/model"
	mail "cloudsweep/notify_services/mail"
	"fmt"
	"html/template"
	"sort"
)

const MAX_RESOURCES_PER_RESOURCE_CLASS = 5

type Sender interface {
	//Send(from string, to []string, subject, body string, isHTML bool) error
	Send(emailDetails mail.EmailDetails) error
}

type EmailManager struct {
	sender Sender
}

func NewEmailManager(sender Sender) *EmailManager {
	return &EmailManager{sender: sender}
}

func NewDefaultEmailManager() *EmailManager {
	return &EmailManager{sender: mail.NewGomailSender(config.GetConfig().Notifications.Email.Host, config.GetConfig().Notifications.Email.Port,
		config.GetConfig().Notifications.Email.Username, config.GetConfig().Notifications.Email.Password)}
}

func (em *EmailManager) SendEmail(emailDetails mail.EmailDetails) error {
	return em.sender.Send(emailDetails)
}

func (em *EmailManager) SendNotification(details model.NotfifyDetails) error {
	logging.NewDefaultLogger().Debugf("Processing the Notification from the channel")
	body, err := buildEmailBody(details)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Problem in building the email body %v", err)
	}
	// Create a CSV file with header and details
	csvData, err := createAttachmentDataInCsvFormat(details.ResourceDetails)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error creating CSV file: %v", err)
		return err
	}
	err = em.SendEmail(mail.EmailDetails{
		To:             details.EmailDetails.ToAddresses,
		From:           config.GetConfig().Notifications.Email.FromAddress,
		Subject:        "CloudSweeper Resource Usage Notification",
		BodyHTML:       body,
		ImageLocations: []string{config.GetConfig().CouldSweeper.Logo},
		AttachmentName: "resource_details.csv",
		AttachmentData: []byte(csvData),
	})
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error: %v", err)
	}
	return err
}

func buildEmailBody(details model.NotfifyDetails) (string, error) {
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
					max-width: 300px;
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

				.internal-table tr {
					border-bottom: 0.5px solid #dddddd; /* Border for rows */
				}
			</style>
		</head>

		<body>
			<tr>
				<td colspan="2" style="text-align: center;">
					<img src="cid:cs.jpg" alt="Company Logo" class="logo">
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

				<!-- Iterate over each ResourceClass group -->
				{{range $resourceClass, $resources := .GroupedResources}}
				<tr>
					<td colspan="2" style="vertical-align: bottom;">
						Resource Class: <strong style="margin-bottom: 0;">{{ $resourceClass }}</strong>
						<table class="internal-table">
							<tr>
								<th>Account ID</th>
								<th>Resource ID</th>
								<th>Resource Name</th>
								<th>Region Code</th>
								<th>Monthly Price</th>
								<th>Current Resource Type</th>
								<th>Recommended Resource Type</th>
								<th>Estimated Monthly Savings</th>
							</tr>
							{{range $index, $resource := $resources}}
							<tr>
								<td>{{.AccountID}}</td>
								<td>{{.ResourceId}}</td>
								<td>{{.ResourceName}}</td>
								<td>{{.RegionCode}}</td>
								<td>{{printf "$%.2f" .MonthlyPrice}}</td>
								<td>{{.CurrentResourceType}}</td>
								<td>{{.RecommendedResourceType}}</td>
								<td>{{printf "$%.2f" .MonthlySavings}}</td>
							</tr>
							{{end}}
						</table>
					</td>
				</tr>
				{{end}}
			</table>

			<div style="text-align: center; margin-top: 20px;">
				<a href="{{.CompanyURL}}" class="button" style="background-color: #3b6696; color: white;" target="_blank">Visit CloudSweeper</a>
			</div>
		</body>
	</html>`

	// Prepare data for the template
	data := struct {
		//LogoURL             string
		CompanyURL          string
		ResourceDetails     []model.NotifyResourceDetails
		PipeLineName        string
		TotalMonthlyPrice   float64
		TotalMonthlySavings float64
		GroupedResources    map[string][]model.NotifyResourceDetails
	}{
		//LogoURL:             logoURL,
		CompanyURL:          config.GetConfig().CouldSweeper.URL,
		ResourceDetails:     details.ResourceDetails,
		PipeLineName:        details.PipeLineName,
		TotalMonthlyPrice:   totalMonthlyPrice,
		TotalMonthlySavings: totalMonthlySavings,
		GroupedResources:    groupByResourceClass(details.ResourceDetails),
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

func groupByResourceClass(resources []model.NotifyResourceDetails) map[string][]model.NotifyResourceDetails {
	groups := make(map[string][]model.NotifyResourceDetails)

	for _, resource := range resources {
		group := groups[resource.ResourceClass]

		// Sort the group based on MonthlySavings in descending order
		sort.Slice(group, func(i, j int) bool {
			return group[i].MonthlySavings > group[j].MonthlySavings
		})

		if len(group) < MAX_RESOURCES_PER_RESOURCE_CLASS {
			groups[resource.ResourceClass] = append(group, resource)
		}
	}
	return groups
}

// Function to create a CSV file from NotifyResourceDetails
func createAttachmentDataInCsvFormat(resources []model.NotifyResourceDetails) (string, error) {
	var csvDataBuffer bytes.Buffer

	// Write header to CSV
	csvDataBuffer.WriteString("Account ID,Resource Type,Resource ID,Resource Name,Region Code,Monthly Price,Current Resource Type,Recommended Resource Type, Monthly Savings\n")

	// Write details to CSV
	for _, resource := range resources {
		csvDataBuffer.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%.2f,%s,%s,%.2f\n",
			resource.AccountID,
			resource.ResourceClass,
			resource.ResourceId,
			resource.ResourceName,
			resource.RegionCode,
			resource.MonthlyPrice,
			resource.CurrentResourceType,
			resource.RecommendedResourceType,
			resource.MonthlySavings,
		))
	}

	return csvDataBuffer.String(), nil
}
