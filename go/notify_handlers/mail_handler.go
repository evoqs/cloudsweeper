package notifications

import (
	"bytes"
	"cloudsweep/config"
	logger "cloudsweep/logging"
	logging "cloudsweep/logging"
	model "cloudsweep/notify_handlers/model"
	mail "cloudsweep/notify_services/mail"
	"fmt"
	"html/template"
	"os"
	"reflect"
	"sort"
	"time"
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
	logging.NewDefaultLogger().Debugf("Body: " + body)
	// Create a CSV file with header and details
	csvData, err := createAttachmentDataInCsvFormat(details.ResourceDetails)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Error creating CSV file: %v", err)
		return err
	}
	// Check if the logo file exists
	var logoPaths []string
	if _, err := os.Stat(config.GetConfig().CouldSweeper.Logo); os.IsNotExist(err) {
		logging.NewDefaultLogger().Errorf("Logo file does not exist: %v", err)
	} else {
		logoPaths = []string{config.GetConfig().CouldSweeper.Logo}
	}
	err = em.SendEmail(mail.EmailDetails{
		To:             details.EmailDetails.ToAddresses,
		From:           config.GetConfig().Notifications.Email.FromAddress,
		Subject:        "CloudSweeper: Resource Usage Notification for Account " + details.AccountId,
		BodyHTML:       body,
		ImageLocations: logoPaths,
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
					max-width: 200px;
					max-height: 100px;
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
					margin-top: 50px; /* Add margin-top to move the table down */
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
					padding-top: 0px
				}

				.internal-table th {
					background-color: #e8e8e8;
				}

				.internal-table tr {
					border-bottom: 0.5px solid #dddddd; /* Border for rows */
				}
				/* Note styling */
				.note-section {
					text-align: right;
					margin-top: 0px;
					font-size: 10px;
					color: #888888;
				}

				.note-icon {
					font-weight: bold;
					margin-right: 5px;
				}

				/* Additional styling for the CSV message */
				.csv-message {
					text-align: center;
					margin-top: 20px;
					font-size: 14px;
					color: #888888;
				}

				.message-row {
					background-color: #d4d4e2;
				}
			</style>
		</head>

		<body>
			<tr></tr>
			<table class="external-table">
				<tr class="external-table-heading">
					<td style="text-align: center;">
						<h1>Resources Summary</h1>
					</td>
					<td style="text-align: center;">
					<img src="cid:logo.jpeg" alt="Company Logo" class="logo">
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
					<td style="text-align: left;"> 
						<span>Account ID:</span>
						<h3 style="margin: 0;">{{.AccountId}}</h3>
					</td>
					<td style="text-align: right;">
						<span>Monthly Savings:</span>
						<h3 style="margin: 0;">{{printf "$%.2f" .TotalMonthlySavings}}</h3>
					</td>
				</tr>
				<tr>
					<td colspan="2"><!-- <span style="font-size: 24px;">&#128339;</span> -->{{formatTime .Time}} </td>
				</tr>

				<!-- Check for no resources - custom message -->
				{{if eq (len .GroupedResources) 0}}
					<tr class="message-row">
						<td colspan="2" style="text-align: center;">
							<!-- shrug 129335 -->
							<span style="font-size: 24px;">&#128269;</span>
							<h3>No resource found for this pipeline</h3>
						</td>
					</tr>
				{{end}}

				<!-- Iterate over each ResourceClass group -->
				{{range $resourceClass, $groupResourceDetails := .GroupedResources}}
				<tr style="height: 55px; vertical-align: bottom;">
					<!-- Resource Class cell -->
					<td style="vertical-align: bottom; padding-bottom: 0px">
						Resource Class: <strong>{{ $resourceClass }}</strong>
					</td>
					<!-- Note cell &#9432-->
					<td style="vertical-align: bottom; padding-bottom: 0px">
						<div class="note-section">
							<span class="note-icon">&#9432;</span>Note: Tables only display top 5 resources based on Monthly Savings
						</div>
					</td>
				</tr>
				<tr style="height: 0px; vertical-align: top; padding-top: 0px;">
					<td colspan="2" style="vertical-align: top; padding-top: 0px">
						<table class="internal-table">
							<tr>
								<!-- <th>Account ID</th> -->
								{{if $groupResourceDetails.Columns.ResourceId}}<th>Resource Id</th>{{end}}
								{{if $groupResourceDetails.Columns.ResourceName}}<th>Resource Name</th>{{end}}
								{{if $groupResourceDetails.Columns.RegionCode}}<th>Region Code</th>{{end}}
								{{if $groupResourceDetails.Columns.MonthlyPrice}}<th>Monthly Price</th>{{end}}
								{{if $groupResourceDetails.Columns.CurrentResourceType}}<th>Resource Type</th>{{end}}
								{{if $groupResourceDetails.Columns.Recommendation}}<th>Recommendation</th>{{end}}
								{{if $groupResourceDetails.Columns.MonthlySavings}}<th>Estimated Monthly Savings</th>{{end}}
							</tr>
							{{range $index, $resource := $groupResourceDetails.Resources}}
							<tr>
								<!-- <td>{{.AccountID}}</td> -->
								{{if $groupResourceDetails.Columns.ResourceId}}<td>{{.ResourceId}}</td>{{end}}
								{{if $groupResourceDetails.Columns.ResourceName}}<td>{{.ResourceName}}</td>{{end}}
								{{if $groupResourceDetails.Columns.RegionCode}}<td>{{.RegionCode}}</td>{{end}}
								{{if $groupResourceDetails.Columns.MonthlyPrice}}<td>{{printf "$%.2f" .MonthlyPrice}}</td>{{end}}
								{{if $groupResourceDetails.Columns.CurrentResourceType}}<td>{{.CurrentResourceType}}</td>{{end}}
								{{if $groupResourceDetails.Columns.Recommendation}}<td>{{.Recommendation}}</td>{{end}}
								{{if $groupResourceDetails.Columns.MonthlySavings}}<td>{{printf "$%.2f" .MonthlySavings}}</td>{{end}}
							</tr>
							{{end}}
						</table>
					</td>
				</tr>
				{{end}}
			</table>

			<!-- CSV message section -->
			<tr>
				<td colspan="2" class="csv-message">
				<span style="font-size: 16px; margin-right: 5px;">&#128206;</span>Please check the attached CSV file for the full list of resources and their details.
				</td>
			</tr>
			<div style="text-align: center; margin-top: 20px;">
				<a href="{{.CompanyURL}}" class="button" style="background-color: #3b6696; color: white;" target="_blank">Visit CloudSweeper</a>
			</div>

			<!-- Terms and Conditions section -->
			<tr>
				<td colspan="2" style="text-align: center; margin-top: 20px;">
					<p style="font-size: 12px; color: #888888;">
						&copy; 2024 CloudSweeper. All rights reserved. | Mailing Address: sales@cloudsweeper.com | 
						<a href="https://cloudsweeper.in/terms" style="color: #888888; text-decoration: none;" target="_blank">Terms and Conditions</a>
					</p>
				</td>
			</tr>

		</body>
	</html>`
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("02 Jan 2006 03:04:05 PM UTC")
		},
	}

	// Prepare data for the template
	data := struct {
		//LogoURL             string
		CompanyURL string
		//ResourceDetails     []model.NotifyResourceDetails
		PipeLineName        string
		AccountId           string
		TotalMonthlyPrice   float64
		TotalMonthlySavings float64
		Time                time.Time
		GroupedResources    map[string]GroupResourceDetails
	}{
		//LogoURL:             logoURL,
		CompanyURL: config.GetConfig().CouldSweeper.URL,
		//ResourceDetails:     details.ResourceDetails,
		PipeLineName:        details.PipeLineName,
		AccountId:           details.AccountId,
		TotalMonthlyPrice:   totalMonthlyPrice,
		TotalMonthlySavings: totalMonthlySavings,
		Time:                details.Time,
		GroupedResources:    groupByResourceClass(details.ResourceDetails),
	}

	// Execute the template
	tmpl, err := template.New("emailTemplate").Funcs(funcMap).Parse(emailTemplate)
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

type GroupResourceDetails struct {
	Resources []model.NotifyResourceDetails
	Columns   map[string]bool
}

func groupByResourceClass(resources []model.NotifyResourceDetails) map[string]GroupResourceDetails {
	logger.NewDefaultLogger().Debugf("Running groupByResourceClass with resource length: %d", len(resources))
	allGroupMap := make(map[string]GroupResourceDetails)

	for _, resource := range resources {
		//logger.NewDefaultLogger().Debugf("For - Resource Class: %s", resource.ResourceClass)
		groupMap := allGroupMap[resource.ResourceClass]
		group := groupMap.Resources

		//logger.NewDefaultLogger().Debugf("Adding Resource: %+v", resource)
		groupMap.Resources = append(group, resource)

		// Sort the group based on MonthlySavings in descending order.
		sort.Slice(group, func(i, j int) bool {
			if group[i].MonthlySavings != group[j].MonthlySavings {
				return group[i].MonthlySavings > group[j].MonthlySavings
			}
			// If MonthlySavings is the same, sort based on MonthlyPrice.
			return group[i].MonthlyPrice > group[j].MonthlyPrice
		})
		if len(groupMap.Resources) > MAX_RESOURCES_PER_RESOURCE_CLASS {
			groupMap.Resources = groupMap.Resources[:MAX_RESOURCES_PER_RESOURCE_CLASS]
		}
		allGroupMap[resource.ResourceClass] = groupMap
	}
	//logger.NewDefaultLogger().Debugf("Length of the allGroupMap: %d", len(allGroupMap))
	for resourceClass, groupResourceDetails := range allGroupMap {
		if groupResourceDetails.Columns == nil {
			groupResourceDetails.Columns = make(map[string]bool)
		}
		for field, isEmpty := range checkEmptyFields(groupResourceDetails.Resources) {
			logger.NewDefaultLogger().Debugf("Field %s isEmpty %v", field, isEmpty)
			groupResourceDetails.Columns[field] = !isEmpty
		}
		allGroupMap[resourceClass] = groupResourceDetails
		//logger.NewDefaultLogger().Debugf("Class %s Column %+v", resourceClass, groupResourceDetails.Columns)
	}
	return allGroupMap
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
			resource.Recommendation,
			resource.MonthlySavings,
		))
	}

	return csvDataBuffer.String(), nil
}

func checkEmptyFields(resources []model.NotifyResourceDetails) map[string]bool {
	logger.NewDefaultLogger().Debugf("Running checkEmptyFields")
	emptyFields := make(map[string]bool)

	resourceType := reflect.TypeOf(resources).Elem()
	fieldsToCheck := getFields(resourceType)

	for _, field := range fieldsToCheck {
		emptyFields[field] = false // Initialize to false, assuming all are empty initially

		// Iterate through the group and check if any value for the field is non-empty
		for _, resource := range resources {
			// Returns true if empty
			if isEmptyField(resource, field) {
				emptyFields[field] = true
				break
			}
		}
	}
	return emptyFields
}

func getFields(resourceType reflect.Type) []string {
	var fieldsToCheck []string

	for i := 0; i < resourceType.NumField(); i++ {
		fieldsToCheck = append(fieldsToCheck, resourceType.Field(i).Name)
	}
	return fieldsToCheck
}

func isEmptyField(resource model.NotifyResourceDetails, field string) bool {
	val := reflect.ValueOf(resource)
	fieldVal := reflect.Indirect(val).FieldByName(field)

	switch fieldVal.Kind() {
	case reflect.String:
		return fieldVal.String() == ""
	case reflect.Float64, reflect.Float32:
		return fieldVal.Float() == 0.0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fieldVal.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fieldVal.Uint() == 0
	// TODO: Not interested in other types for now
	/*case reflect.Bool:
		return !fieldVal.Bool()
	case reflect.Slice, reflect.Array, reflect.Map:
		return fieldVal.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return fieldVal.IsNil()*/
	default:
		return false
	}
}
