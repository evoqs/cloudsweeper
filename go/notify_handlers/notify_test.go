package notifications

import (
	"fmt"
	"testing"
	"time"

	"cloudsweep/config"
	"cloudsweep/storage"
	"cloudsweep/utils"
)

func TestNotifications(t *testing.T) {
	cfg := config.LoadConfig()
	dbUrl, err := utils.GetDBUrl(&cfg)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(dbUrl)
	}

	dbM := storage.GetDBManager()
	dbM.SetDbUrl(dbUrl)
	dbM.SetDatabase(config.GetConfig().Database.Name)

	_, err = dbM.Connect()
	if err != nil {
		fmt.Println("Failed to connect to DB " + err.Error())
	}
	err = dbM.CheckConnection()
	if err != nil {
		fmt.Println("Connection Check failed")
	} else {
		fmt.Println("Successfully Connected")
		defer dbM.Disconnect()
	}
	storage.MakeDBOperators(dbM)
	StartNotificationService()

	/*details := notify_model.NotfifyDetails{
		// Populate the details as needed
		PipeLineName: "Cut Cost",
		ResourceDetails: []notify_model.NotifyResourceDetails{
			{
				AccountID:     "123456789",
				ResourceClass: "ComputeInstance",
				ResourceId:    "i-abcdef1234567890",
				ResourceName:  "MyInstance",
				RegionCode:    "us-east-1",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "John Doe",
				},
				MonthlyPrice:            8.78,
				CurrentResourceType:     "t2.large",
				RecommendedResourceType: "t2.micro",
				MonthlySavings:          0.34,
			},
			{
				AccountID:     "1234567890",
				ResourceClass: "ComputeInstance",
				ResourceId:    "i-abcdef12345678901",
				ResourceName:  "MyInstance2",
				RegionCode:    "us-east-1",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "Pavan K",
				},
				MonthlyPrice:            143.4,
				CurrentResourceType:     "mx7.large",
				RecommendedResourceType: "mx4.large",
				MonthlySavings:          2.87,
			},
			{
				AccountID:     "1234567890",
				ResourceClass: "EBS",
				ResourceId:    "i-abcdef12345678901",
				ResourceName:  "MyInstance2",
				RegionCode:    "us-east-1",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "Pavan K",
				},
				MonthlyPrice:            12.4,
				CurrentResourceType:     "mx7.large",
				RecommendedResourceType: "mx4.large",
				MonthlySavings:          0.87,
			},
			{
				AccountID:     "1234567890",
				ResourceClass: "EBS",
				ResourceId:    "i-abcdef12345678901",
				ResourceName:  "MyInstance2",
				RegionCode:    "us-east-1",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "Pavan K",
				},
				MonthlyPrice:            34.7,
				CurrentResourceType:     "mx7.large",
				RecommendedResourceType: "mx4.large",
				MonthlySavings:          0.98,
			},
			{
				AccountID:     "1234567890",
				ResourceClass: "ComputeInstance",
				ResourceId:    "i-abcdef12345678901",
				ResourceName:  "MyInstance59",
				RegionCode:    "us-east-1",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "Pavan K",
				},
				MonthlyPrice:            143.3,
				CurrentResourceType:     "mx7.large",
				RecommendedResourceType: "mx4.large",
				MonthlySavings:          0.7843,
			},
			{
				AccountID:     "1234567890",
				ResourceClass: "ComputeInstance",
				ResourceId:    "i-abcdef12345678901",
				ResourceName:  "MyInstance45",
				RegionCode:    "us-east-1",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "Pavan K",
				},
				MonthlyPrice:            100,
				CurrentResourceType:     "mx7.large",
				RecommendedResourceType: "mx4.large",
				MonthlySavings:          6.9,
			},
			{
				AccountID:     "1234567890",
				ResourceClass: "ComputeInstance",
				ResourceId:    "i-dsada12345678901",
				ResourceName:  "MyInstance89",
				RegionCode:    "us-east-1",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "Pavan K",
				},
				MonthlyPrice:            100,
				CurrentResourceType:     "mx7.large",
				RecommendedResourceType: "mx4.large",
				MonthlySavings:          0.34,
			},
			{
				AccountID:     "1234567890",
				ResourceClass: "ComputeInstance",
				ResourceId:    "i-32874672384782",
				ResourceName:  "MyInstance90",
				RegionCode:    "us-east-5",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "Pavan K",
				},
				MonthlyPrice:            100,
				CurrentResourceType:     "mx7.large",
				RecommendedResourceType: "mx4.large",
				MonthlySavings:          0.00,
			},
			{
				AccountID:     "1234567890",
				ResourceClass: "ComputeInstance",
				ResourceId:    "i-dsada12345678901",
				ResourceName:  "MyInstance89",
				RegionCode:    "us-east-1",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "Pavan K",
				},
				MonthlyPrice:        45.77,
				CurrentResourceType: "mx7.large",
			},
		},
		EmailDetails: notify_model.NotifyEmailDetails{
			Enabled:     true,
			ToAddresses: []string{"pavan.vitla@gmail.com", "bipinkmg@gmail.com", "jnagendran78@gmail.com", "web.jlingasur@gmail.com"},
		},
		SlackDetails: notify_model.NotifySlackDetails{
			Enabled: true,
			Url:     []string{"https://slack-webhook-url"},
		},
		WebhookDetails: notify_model.NotifyWebhookDetails{
			Enabled: true,
			Url:     []string{"https://webhook-url"},
		},
	}*/
	// TODO: Get the pipeline ID and send the notification.
	t.Logf("Sending Notification..")
	pipelineId := "65a0b6f154222ca8f93a651f"
	for i := 0; i < 1; i++ {
		fmt.Printf("Sending notification #%d\n", i+1)
		SendNotification(pipelineId)
	}
	time.Sleep(10 * time.Second)
}
