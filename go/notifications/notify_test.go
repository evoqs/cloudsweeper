package notifications

import (
	"fmt"
	"testing"
	"time"

	"cloudsweep/config"
	notify_model "cloudsweep/notifications/model"
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

	details := notify_model.NotfifyDetails{
		// Populate the details as needed
		PipeLineName: "Cut Cost",
		ResourceDetails: []notify_model.NotifyResourceDetails{
			{
				AccountID:    "123456789",
				ResourceType: "ComputeInstance",
				ResourceId:   "i-abcdef1234567890",
				ResourceName: "MyInstance",
				RegionCode:   "us-east-1",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "John Doe",
				},
				MonthlyPrice:   10,
				Recommendation: "t2.micro",
				MonthlySavings: 0.34,
			},
			{
				AccountID:    "1234567890",
				ResourceType: "ComputeInstance",
				ResourceId:   "i-abcdef12345678901",
				ResourceName: "MyInstance2",
				RegionCode:   "us-east-1",
				ResourceTags: map[string]string{
					"Environment": "Production",
					"Owner":       "Pavan K",
				},
				MonthlyPrice:   100,
				Recommendation: "mx4.large",
				MonthlySavings: 2.87,
			},
		},
		EmailDetails: notify_model.NotifyEmailDetails{
			Enabled:     true,
			ToAddresses: []string{"pavan.vitla@gmail.com", "bipinkmg@gmail.com"}, //, "jnagendran78@gmail.com", "web.jlingasur@gmail.com"},
		},
		/*SlackDetails: notify_model.NotifySlackDetails{
			Enabled: true,
			Url:     []string{"https://slack-webhook-url"},
		},
		WebhookDetails: notify_model.NotifyWebhookDetails{
			Enabled: true,
			Url:     []string{"https://webhook-url"},
		},*/
	}
	t.Logf("Sending Notification..")
	for i := 0; i < 1; i++ {
		fmt.Printf("Sending notification #%d\n", i+1)
		SendNotification(details)
	}
	time.Sleep(10 * time.Second)
}
