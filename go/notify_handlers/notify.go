package notifications

import (
	logging "cloudsweep/logging"
	aws_model "cloudsweep/model/aws"
	notify_model "cloudsweep/notify_handlers/model"
	"cloudsweep/storage"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	notifiers         map[string]Notifier
	pipeLineIdChannel chan string
}

// NewNotifyManager initializes a new NotifyManager.
func newNotifyManager() *NotifyManager {
	return &NotifyManager{
		notifiers:         make(map[string]Notifier),
		pipeLineIdChannel: make(chan string),
	}
}

// RegisterNotifier registers a notifier for a specific channel.
func (nm *NotifyManager) RegisterNotifier(channel NotificationChannel, notifier Notifier) {
	nm.notifiers[string(channel)] = notifier // Change from append to direct assignment
}

// SendNotification sends a notification using the specified channels.
func (nm *NotifyManager) SendNotification(pipelineId string) {
	logging.NewDefaultLogger().Debugf("Adding Pipeline Id to teh channel")
	nm.pipeLineIdChannel <- pipelineId
}

// StartProcessing starts the processing loop for notifications.
func (nm *NotifyManager) StartProcessing() {
	go func() {
		for {
			select {
			case request := <-nm.pipeLineIdChannel:
				fmt.Println("\n---------> PAVAN \n", request)
				go nm.processNotification(request)
			}
		}
	}()
}

func (nm *NotifyManager) processNotification(pipelineId string) {
	logging.NewDefaultLogger().Debugf("Processing the pipelineId from the channel")
	request, err := processPipelineResult(pipelineId)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Skipping the Notification for pipeline: %s Reason: %v", pipelineId, err)
		return
	}
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
	close(notifyManager.pipeLineIdChannel)
}

func SendNotification(pipelineId string) {
	if notifyManager != nil {
		notifyManager.SendNotification(pipelineId)
	} else {
		logging.NewDefaultLogger().Error("Notification service is not started. Call StartNotificationService() first.")
	}
}

// Owner: Bibin
func processPipelineResult(pipeLineId string) (notify_model.NotfifyDetails, error) {
	opr := storage.GetDefaultDBOperators()
	query := fmt.Sprintf(`{"pipelineid": "%s"}`, pipeLineId)
	pipeline, err := opr.PipeLineOperator.GetPipeLineDetails(pipeLineId)
	if err != nil {
		logging.NewDefaultLogger().Errorf("Failed to get the pipeline details from DB: %v", err)
	}
	results, _ := opr.PolicyOperator.GetPolicyResultDetails(query)

	// TODO: check if pipeline is not empty
	cloudAccList, err := opr.AccountOperator.GetCloudAccount(pipeline[0].CloudAccountID)
	if err != nil || len(cloudAccList) < 1 {
		logging.NewDefaultLogger().Errorf("Failed to get cloundaccount details for CloudAccountId %s, %s", pipeline[0].CloudAccountID, err.Error())
	}

	awsAccountID := cloudAccList[0].AwsCredentials.AccountID
	var details notify_model.NotfifyDetails

	details.EmailDetails.ToAddresses = pipeline[0].Notification.EmailAddresses
	// TODO: Bibin - This info should be fetched from UI, set in Pipeline and propogated here
	details.EmailDetails.Enabled = true
	var error error
	line := "----------------------------------------------"
	for _, object := range results {
		var resource notify_model.NotifyResourceDetails
		resource.AccountID = awsAccountID

		if object.Resource == "ec2" {
			for _, result := range object.Resultlist {
				if result.Result == nil {
					continue
				}
				resultList, ok := result.Result.(primitive.A)
				if !ok {
					fmt.Println("Invalid result set")
					continue
				}

				var data aws_model.AwsInstanceResult
				//ec2Result := result.Result.(*[]aws_model.AwsInstanceResult)
				for _, entry := range resultList {

					primativeData := entry.(primitive.D)
					tempByteHolder, _ := bson.MarshalExtJSON(primativeData, true, true)
					bson.UnmarshalExtJSON(tempByteHolder, true, &data)
					fmt.Printf("%s\n ----->  %v\n%s", line, data, line)
					resource.CurrentResourceType = data.ResultData.InstanceType
					// TODO: Bibin - Monthly Price and Monthly savings should be provided with float value with separate currency and metric
					if data.MetaData != nil {
						resource.MonthlyPrice, error = strconv.ParseFloat(strings.Split(data.MetaData.Cost, " ")[0], 64)
						if error != nil {
							logging.NewDefaultLogger().Errorf("Error While converting Monthly Price: %v", error)
						}

						if data.MetaData.Recommendations != nil {
							resource.MonthlySavings, err = strconv.ParseFloat(strings.Split(data.MetaData.Recommendations[0].EstimatedCostSavings, " ")[0], 64)
							if error != nil {
								logging.NewDefaultLogger().Errorf("Error While converting Monthly Savings: %v", error)
							}
							resource.RecommendedResourceType = data.MetaData.Recommendations[0].Recommendation
						} else {
							resource.MonthlySavings = 0.0
							resource.RecommendedResourceType = ""
						}
					}

					resource.RegionCode = result.Region
					resource.ResourceClass = "Compute Instances"
					resource.ResourceId = data.ResultData.InstanceId
					//resource.ResourceName = ""
					//resource.ResourceTags = ""
					details.ResourceDetails = append(details.ResourceDetails, resource)
				}
			}
		} else if object.Resource == "ebs" {
			for _, result := range object.Resultlist {
				if result.Result == nil {
					continue
				}
				resultList, ok := result.Result.(primitive.A)
				if !ok {
					fmt.Println("Invalid result set")
					continue
				}

				var data aws_model.AwsBlockVolumeResult
				for _, entry := range resultList {

					primativeData := entry.(primitive.D)
					tempByteHolder, _ := bson.MarshalExtJSON(primativeData, true, true)
					bson.UnmarshalExtJSON(tempByteHolder, true, &data)
					fmt.Printf("%s\n ----->  %v\n%s", line, data, line)
					resource.CurrentResourceType = data.ResultData.VolumeType
					// TODO: Bibin - Monthly Price and Monthly savings should be provided with float value with separate currency and metric
					if data.MetaData != nil {
						resource.MonthlyPrice, error = strconv.ParseFloat(strings.Split(data.MetaData.Cost, " ")[0], 64)
						if error != nil {
							logging.NewDefaultLogger().Errorf("Error While converting Monthly Price: %v", error)
						}

						if data.MetaData.Recommendations != nil {
							resource.MonthlySavings, err = strconv.ParseFloat(strings.Split(data.MetaData.Recommendations[0].EstimatedCostSavings, " ")[0], 64)
							if error != nil {
								logging.NewDefaultLogger().Errorf("Error While converting Monthly Savings: %v", error)
							}

							resource.RecommendedResourceType = data.MetaData.Recommendations[0].Recommendation
						} else {
							resource.MonthlySavings = 0.0
							resource.RecommendedResourceType = ""
						}
					}

					resource.RegionCode = result.Region
					resource.ResourceClass = "EBS Volumes"
					resource.ResourceId = data.ResultData.VolumeId
					//resource.ResourceName = ""
					//resource.ResourceTags = ""
					details.ResourceDetails = append(details.ResourceDetails, resource)
				}
			}

		} else if object.Resource == "elastic-ip" {
			for _, result := range object.Resultlist {
				if result.Result == nil {
					continue
				}
				resultList, ok := result.Result.(primitive.A)
				if !ok {
					fmt.Println("Invalid result set")
					continue
				}

				var data aws_model.AwsElasticIPResult
				for _, entry := range resultList {

					primativeData := entry.(primitive.D)
					tempByteHolder, _ := bson.MarshalExtJSON(primativeData, true, true)
					bson.UnmarshalExtJSON(tempByteHolder, true, &data)
					fmt.Printf("%s\n ----->  %v\n%s", line, data, line)
					if data.ResultData.Domain == "vpc" {
						resource.CurrentResourceType = "EC2-VPC"
					} else {
						resource.CurrentResourceType = "EC2-Classic"
					}

					// TODO: Bibin - Monthly Price and Monthly savings should be provided with float value with separate currency and metric
					if data.MetaData != nil {
						resource.MonthlyPrice, error = strconv.ParseFloat(strings.Split(data.MetaData.Cost, " ")[0], 64)
						if error != nil {
							logging.NewDefaultLogger().Errorf("Error While converting Monthly Price: %v", error)
						}
						if data.MetaData.Recommendations != nil {
							resource.MonthlySavings, err = strconv.ParseFloat(strings.Split(data.MetaData.Recommendations[0].EstimatedCostSavings, " ")[0], 64)
							if error != nil {
								logging.NewDefaultLogger().Errorf("Error While converting Monthly Savings: %v", error)
							}

							resource.RecommendedResourceType = data.MetaData.Recommendations[0].Recommendation
						} else {
							resource.MonthlySavings = 0.0
							resource.RecommendedResourceType = ""
						}
					}

					resource.RegionCode = result.Region
					resource.ResourceClass = "Elastic IP"
					resource.ResourceId = "TODO"
					//resource.ResourceName = ""
					//resource.ResourceTags = ""
					details.ResourceDetails = append(details.ResourceDetails, resource)
				}
			}

		} else if object.Resource == "ebs-snapshot" {
			for _, result := range object.Resultlist {
				if result.Result == nil {
					continue
				}
				resultList, ok := result.Result.(primitive.A)
				if !ok {
					fmt.Println("Invalid result set")
					continue
				}

				var data aws_model.AwsSnapshotResult
				for _, entry := range resultList {

					primativeData := entry.(primitive.D)
					tempByteHolder, _ := bson.MarshalExtJSON(primativeData, true, true)
					bson.UnmarshalExtJSON(tempByteHolder, true, &data)
					fmt.Printf("%s\n ----->  %v\n%s", line, data, line)
					resource.CurrentResourceType = data.ResultData.StorageTier

					// TODO: Bibin - Monthly Price and Monthly savings should be provided with float value with separate currency and metric
					if data.MetaData != nil {
						resource.MonthlyPrice, error = strconv.ParseFloat(strings.Split(data.MetaData.Cost, " ")[0], 64)
						if error != nil {
							logging.NewDefaultLogger().Errorf("Error While converting Monthly Price: %v", error)
						}
						if data.MetaData.Recommendations != nil {
							resource.MonthlySavings, err = strconv.ParseFloat(strings.Split(data.MetaData.Recommendations[0].EstimatedCostSavings, " ")[0], 64)
							if error != nil {
								logging.NewDefaultLogger().Errorf("Error While converting Monthly Savings: %v", error)
							}

							resource.RecommendedResourceType = data.MetaData.Recommendations[0].Recommendation
						} else {
							resource.MonthlySavings = 0.0
							resource.RecommendedResourceType = ""
						}
					}

					resource.RegionCode = result.Region
					resource.ResourceClass = "EBS Snapshot"
					resource.ResourceId = data.ResultData.SnapshotId
					//resource.ResourceName = ""
					//resource.ResourceTags = ""
					details.ResourceDetails = append(details.ResourceDetails, resource)
				}
			}

		}
		// TODO: Add the section for elastic IP and EBSSnapshot
	}

	logging.NewDefaultLogger().Debugf("%s\n%+v\n%s", line, details, line)
	details.PipeLineName = pipeline[0].PipeLineName
	return details, error
}
