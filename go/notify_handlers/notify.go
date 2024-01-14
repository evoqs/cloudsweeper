package notifications

import (
	logging "cloudsweep/logging"
	notify_model "cloudsweep/notify_handlers/model"
	"sync"
)

type NotificationChannel string

var (
	notifyManager *NotifyManager
	once          sync.Once
)

const (
	EmailChannel   NotificationChannel = "email"
	SlackChannel   NotificationChannel = "slack"
	WebhookChannel NotificationChannel = "webhook"
	// Add more channels as needed
)

type Notifier interface {
	SendNotification(details notify_model.NotfifyDetails) error
}

type NotifyManager struct {
	notifiers map[string]Notifier
	details   chan notify_model.NotfifyDetails
}

// NewNotifyManager initializes a new NotifyManager.
func newNotifyManager() *NotifyManager {
	return &NotifyManager{
		notifiers: make(map[string]Notifier),
		details:   make(chan notify_model.NotfifyDetails),
	}
}

// RegisterNotifier registers a notifier for a specific channel.
func (nm *NotifyManager) RegisterNotifier(channel NotificationChannel, notifier Notifier) {
	nm.notifiers[string(channel)] = notifier // Change from append to direct assignment
}

// SendNotification sends a notification using the specified channels.
func (nm *NotifyManager) SendNotification(details notify_model.NotfifyDetails) {
	logging.NewDefaultLogger().Debugf("Adding Notification Details to teh channel")
	nm.details <- details
}

// StartProcessing starts the processing loop for notifications.
func (nm *NotifyManager) StartProcessing() {
	go func() {
		for {
			select {
			case request := <-nm.details:
				go nm.processNotification(request)
			}
		}
	}()
}

func (nm *NotifyManager) processNotification(request notify_model.NotfifyDetails) {
	logging.NewDefaultLogger().Debugf("Processing the Notification from the channel")
	var channels []NotificationChannel

	// Check if email details are enabled
	if request.EmailDetails.Enabled {
		channels = append(channels, EmailChannel)
	}

	// Check if slack details are enabled
	if request.SlackDetails.Enabled {
		channels = append(channels, SlackChannel)
	}

	// Check if webhook details are enabled
	if request.WebhookDetails.Enabled {
		channels = append(channels, WebhookChannel)
	}

	for _, channel := range channels {
		notifier, ok := nm.notifiers[string(channel)]
		if !ok {
			logging.NewDefaultLogger().Errorf("Unsupported notification channel: %s\n", channel)
			continue
		}

		logging.NewDefaultLogger().Debugf("Sending Notification via %s", channel)
		err := notifier.SendNotification(request)
		if err != nil {
			logging.NewDefaultLogger().Errorf("Error sending notification via %s: %s\n", channel, err)
		}
	}
}

func StartNotificationService() {
	once.Do(func() {
		logging.NewDefaultLogger().Infof("Starting the Notification Service")
		notifyManager = newNotifyManager()
		notifyManager.RegisterNotifier(EmailChannel, NewDefaultEmailManager())
		go notifyManager.StartProcessing()
	})
}

// StopNotificationService stops the processing of notifications.
func StopNotificationService() {
	logging.NewDefaultLogger().Infof("Stopping the Notification Service")
	close(notifyManager.details)
}

func SendNotification(details notify_model.NotfifyDetails) {
	if notifyManager != nil {
		notifyManager.SendNotification(details)
	} else {
		logging.NewDefaultLogger().Error("Notification service is not started. Call StartNotificationService() first.")
	}
}
