package mail

import (
	"fmt"
	"testing"

	"cloudsweep/config"
	"cloudsweep/storage"
	"cloudsweep/utils"
)

func TestNotificationEmail(t *testing.T) {
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

	fmt.Printf("Host: %s\nPort: %d\nUsername: %s\nPassword %s\n", config.GetConfig().Notifications.Email.Host, config.GetConfig().Notifications.Email.Port, config.GetConfig().Notifications.Email.Username, config.GetConfig().Notifications.Email.Password)
	gomailSender := NewGomailSender(config.GetConfig().Notifications.Email.Host, config.GetConfig().Notifications.Email.Port, config.GetConfig().Notifications.Email.Username, config.GetConfig().Notifications.Email.Password)

	// Create an EmailManager instance with the GomailSender
	emailManager := NewEmailManager(gomailSender)

	to := []string{"pavan.vitla@gmail.com", "bipinkmg@gmail.com"}
	subject := "Cloud Sweeper Sample Test Mail"
	bodyHTML := "<html><body><p>Cloud Sweeper Sample Test Mail. You should see the resource details here in future.</p></body></html>"
	err = emailManager.SendEmail(config.GetConfig().Notifications.Email.FromAddress, to, subject, bodyHTML, true)
	t.Logf("Email Sent %v", err)
}
