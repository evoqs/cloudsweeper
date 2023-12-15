package runner

import (
	"cloudsweep/cloud_lib"
	"cloudsweep/config"
	aws_cost_estimator "cloudsweep/cost_estimator/aws"

	//cost_estimator "cloudsweep/cost_estimator/aws"
	logger "cloudsweep/logging"
	"cloudsweep/model"
	aws_model "cloudsweep/model/aws"
	"cloudsweep/policy_converter"
	"cloudsweep/storage"
	"cloudsweep/utils"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*var (
	pipelineQueue = make(chan model.PipeLine) // Channel for queuing pipeline requests
	pipelineLock  sync.Mutex                  // Mutex for synchronizing access to pipelineCounter
)*/

/*func StartPipelineRunner() {
	go func() {
		logger.NewDefaultLogger().Infof("PipelineQueue Started")
		defer logger.NewDefaultLogger().Infof("PipelineQueue Stopped")
		for {
			select {
			case pipeline := <-pipelineQueue:
				logger.NewDefaultLogger().Infof("PipelineQueue: Received pipeline: %+v\n", pipeline.PipeLineName)
				go runPipeline(pipeline)
			}
		}
	}()
}*/

func ValidateAndRunPipeline(pipelineid string) (int, error) {
	logwriter := logger.NewDefaultLogger()
	opr := storage.GetDefaultDBOperators()
	// TODO: HTTP Errors should be masked
	// TODO: Check if pipeline is enabled / inprogress
	pipeLineList, err := opr.PipeLineOperator.GetPipeLineDetails(pipelineid)
	if err != nil {
		return 500, err
	}
	if len(pipeLineList) == 0 {
		return 404, fmt.Errorf("Pipeline is not found")
	}
	// Why will the pipeline list be greater than 1?
	pipeLine := pipeLineList[0]
	if !pipeLine.Enabled {
		logwriter.Infof("Cannot run Pipeline in diabled state : %s", pipelineid)
		return 409, errors.New("Cannot Run Pipeline in disabled state.")
	}
	//Validate the policy
	policyids := pipeLine.Policies
	if len(policyids) == 0 {
		logwriter.Infof("Cannot run Pipeline %s, no policies defined", pipelineid)
		return 409, errors.New("No policy Defined for Pipeline.")
	}
	policyId := policyids[0]
	policies, err := opr.PolicyOperator.GetPolicyDetails(policyId)
	if err != nil {
		return 500, err
	}
	if len(policies) == 0 {
		return 404, fmt.Errorf("Policy not found with id %s", policyId)
	}

	accounts, err := opr.AccountOperator.GetCloudAccount(pipeLine.CloudAccountID)
	if err != nil {
		return 500, err
	}
	if len(accounts) == 0 {
		return 404, fmt.Errorf("Account not found with id %s", pipeLine.CloudAccountID)
	}
	account := accounts[0]

	if !utils.ValidateAwsCredentials(account.AwsCredentials.AccessKeyID, account.AwsCredentials.SecretAccessKey) {

		return 409, errors.New("Aws Authentication Failed")
	}
	//pipelineQueue <- pipeLine
	go runPipeline(pipeLine)
	return 200, nil
}

func runPipeline(pipeLine model.PipeLine) {
	logwriter := logger.NewDefaultLogger()
	logwriter.Infof("Running the pipeline: %s", pipeLine.PipeLineID.Hex())
	//Step 1
	if pipeLine.RunStatus == model.RUNNING {
		logwriter.Infof("Run is already in progress for pipline %s", pipeLine.PipeLineID.Hex())
		//return
	}
	pipeLine.RunStatus = model.RUNNING
	opr := storage.GetDefaultDBOperators()
	upcount, err := opr.PipeLineOperator.UpdatePipeLine(pipeLine)
	if upcount != 1 {
		logwriter.Errorf("Failed to update pipline run status for pipeline %s, %s", pipeLine.PipeLineID.Hex(), err.Error())
		return
	}

	//TODO Implement Parallel processing of polices
	policyids := pipeLine.Policies
	isPolicyRunFailed := false
	var wg sync.WaitGroup

	pipeLine.LastRunTime = time.Now().Unix()

	cmap := make(map[string]chan bool)

	for _, policyid := range policyids {

		policyList, err := opr.PolicyOperator.GetPolicyDetails(policyid)
		logwriter.Infof("Running pipeline %s policy, %s", pipeLine.PipeLineID.Hex(), policyList[0].PolicyID.Hex())
		if err != nil {
			logwriter.Errorf("Failed to get policy, get DB operation failed, %s", err.Error())
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Internal DB Error", time.Now().Unix(), nil, false)
			isPolicyRunFailed = true
			continue
		}

		if len(policyList) == 0 {
			logwriter.Errorf("Policy not found for pipline %s pilocyid %s", pipeLine.PipeLineID.Hex(), policyList[0].PolicyID.Hex())
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Policy Definition missing", time.Now().Unix(), nil, false)
			isPolicyRunFailed = true
			continue
		}

		wg.Add(1)
		retunChan := make(chan bool, 1)
		cmap[policyList[0].PolicyID.Hex()] = retunChan
		go runPolicy(&wg, policyList[0], pipeLine, retunChan)
		logwriter.Infof("Starting pipeline %s policy run %s", pipeLine.PipeLineID.Hex(), policyList[0].PolicyID.Hex())

	}
	wg.Wait()
	logwriter.Infof("Waiting for all policies execution to end for pipeline %s", pipeLine.PipeLineID.Hex())

	for elem := range cmap {

		runResult := <-cmap[elem]
		close(cmap[elem])
		logwriter.Infof("Completed pipeline %s policy run %s with result %t", pipeLine.PipeLineID.Hex(), elem, !runResult)

		if runResult {
			isPolicyRunFailed = true
		}

	}

	//

	if isPolicyRunFailed {
		pipeLine.RunStatus = model.FAILED
		upcount, err := opr.PipeLineOperator.UpdatePipeLine(pipeLine)
		if upcount != 1 {
			fmt.Println("Failed to update pipline run status,", err)
			return
		}
	} else {
		pipeLine.RunStatus = model.COMPLETED
		upcount, err := opr.PipeLineOperator.UpdatePipeLine(pipeLine)
		if upcount != 1 {
			fmt.Println("Failed to update pipline run status,", err)
			return
		}
	}
	// TODO: How do you inform back to the UI the reason for failure. pipeline model should be updated to have the reason for last failure and number of previous failures
}

func runPolicy(wg *sync.WaitGroup, policy model.Policy, pipeLine model.PipeLine, rchan chan bool) {
	defer wg.Done()

	logwriter := logger.NewDefaultLogger()
	policyJson := policy.PolicyDefinition
	policyid := policy.PolicyID.Hex()
	opr := storage.GetDefaultDBOperators()
	isPolicyRunFailed := false

	/*Steps involved in running
	1. Validating a run is not already in progress and is Enabled
	2. create a run folder for policy in /tmp
	3. fecthing the policy json and converting into yaml
	4. Fectch the cloud credentails for the policy run
	5. trigger the run and update the pipleline status as running
	6. Send Accepted response 200
	7. Update the results
	8. Mark as Completed
	9. Delete the folder
	*/

	//Step 2 create runfolder
	RunFolder := fmt.Sprintf("/tmp/%s.%s", policyid, strconv.Itoa(rand.Intn(100000)))
	os.Mkdir(RunFolder, os.ModePerm)
	defer os.RemoveAll(RunFolder)

	//Step 3
	policyFile := fmt.Sprintf("%s/%s", RunFolder, "policy.yml")
	err := policy_converter.ConvertJsonToYamlAndWriteToFile(policyJson, policyFile)
	if err != nil {
		logwriter.Errorf("Failed to convert json policy to yaml, for policy id %s, %s", policy.PolicyID, err.Error())
		updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Invalid policy definition", time.Now().Unix(), nil, false)
		rchan <- true
		return
	}

	//Get creds
	cloudAccList, err := opr.AccountOperator.GetCloudAccount(pipeLine.CloudAccountID)
	if err != nil || len(cloudAccList) < 1 {
		logwriter.Errorf("Failed to get cloundaccount details for policy id %s %s , %s", policy.PolicyID, pipeLine.CloudAccountID, err.Error())
		updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Missing Cloud Account definition for policy", time.Now().Unix(), nil, false)
		rchan <- true
		return
	}

	cloudAcc := cloudAccList[0]
	//validating cloud authentication

	if cloudAcc.AccountType == model.AWS {
		awsClient, err := cloud_lib.GetAwsClient(cloudAcc.AwsCredentials.AccessKeyID, cloudAcc.AwsCredentials.SecretAccessKey, "")
		if err != nil {
			logwriter.Errorf("Failed to connect to AWS Client")
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Failed to connect to AWS Client", time.Now().Unix(), nil, false)
			rchan <- true
			return
		}

		if !awsClient.ValidateCredentials() {
			logwriter.Errorf("AWS Authentication failed for cloud account %s", cloudAcc.CloudAccountID.Hex())
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "AWS Authentication Failed, invalid accesskey/secret", time.Now().Unix(), nil, false)
			rchan <- true
			return

		}

		isSingleRegionExecution := true
		if len(pipeLine.ExecutionRegions) > 1 {
			isSingleRegionExecution = false
		} else if pipeLine.ExecutionRegions[0] == "all" {
			isSingleRegionExecution = false
		}

		if isSingleRegionExecution {
			logger.NewDefaultLogger().Logger.Info("Execution for single region")
		}

		//
		regionFlag := utils.ConstructRegionList(&pipeLine.ExecutionRegions)
		var envvars []string
		//envvars = append(envvars, fmt.Sprintf("AWS_DEFAULT_REGION=%s", strings.Join(policy.ExecutionRegions, ",")))
		envvars = append(envvars, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", cloudAcc.AwsCredentials.AccessKeyID))
		envvars = append(envvars, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", cloudAcc.AwsCredentials.SecretAccessKey))

		policyRunTime := time.Now().Unix()
		activatePath := config.GetConfig().Custodian.C7nAwsInstall
		c := make(chan string, 1)
		go utils.RunCustodianPolicy(envvars, RunFolder, policyFile, activatePath, regionFlag, c)
		runres, ok := <-c
		close(c)
		if !ok {
			logwriter.Errorf("Failed to read from channel for policy execution %s", policyid)
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Internal Error", policyRunTime, nil, false)
			rchan <- true
			return

		}

		if strings.Contains(strings.ToUpper(runres), "ERROR") {
			fmt.Println("policy run failed with result", runres)
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Internal Error", policyRunTime, nil, false)
			rchan <- true
			return
		} else {
			logger.NewDefaultLogger().Infof("policy run successful with result %s", runres)
			folderList := utils.GetFolderList(RunFolder)
			var resultList = make([]model.RegionResult, 0)
			resourceName := utils.GetResourceName(policyFile)

			//create Wait group for each results
			var resultWg sync.WaitGroup
			for _, element := range folderList {
				var regionName, resourceFile, policyName string
				if isSingleRegionExecution {
					regionName = pipeLine.ExecutionRegions[0]
					resourceFile = fmt.Sprintf("%s/%s/%s", RunFolder, element, "resources.json")
					policyName = element
					fmt.Println("Single Region", element)
				} else {
					regionName = element
					policyFolder := utils.GetFolderList(fmt.Sprintf("%s/%s", RunFolder, element))
					if len(policyFolder) != 1 {
						logger.NewDefaultLogger().Errorf("Invalid number of folders in c7n multi region execution folder. %v", policyFolder)
						isPolicyRunFailed = true
						continue
					}
					policyName = policyFolder[0]
					resourceFile = fmt.Sprintf("%s/%s/%s/%s", RunFolder, element, policyName, "resources.json")
				}

				logger.NewDefaultLogger().Infof("Reading resource file %s for region %s", resourceFile, regionName)

				regionResult := new(model.RegionResult)

				//get policy name and resource
				resourceList, err := utils.ReadFile(resourceFile)
				replacer := strings.NewReplacer("\r", "", "\n", "")
				resourceList = replacer.Replace(string(resourceList))

				//var out []byte
				//out = []byte("")

				if err != nil {
					logwriter.Errorf("Failed to read policy result from result %s, Error %s", resourceFile, err.Error())
					isPolicyRunFailed = true
					continue
				} else {
					logwriter.Infof("Successfully completed the policy run, %s", policyName)
				}

				//Converting into shorter json
				var IList interface{}
				if resourceName == "ec2" {
					var rList []aws_model.AwsInstanceResult
					var policyresultList []aws_model.AwsInstancePolicyResultData
					json.Unmarshal([]byte(resourceList), &policyresultList)
					for _, elem := range policyresultList {
						var resultData aws_model.AwsInstanceResultData
						resultData.AvailabilityZone = elem.Placement.AvailabilityZone
						resultData.Code = elem.State.Code
						resultData.GroupName = elem.Placement.GroupName
						resultData.InstanceId = elem.InstanceId
						resultData.InstanceType = elem.InstanceType
						resultData.PlatformDetails = elem.PlatformDetails
						resultData.Region = elem.Region
						resultData.StateName = elem.State.Name
						resultData.Tenancy = elem.Placement.Tenancy

						var resultEntry aws_model.AwsInstanceResult
						resultEntry.ResultData = resultData
						resultEntry.MetaData = nil

						if policy.PolicyType == "Default" {
							var metaData model.ResultMetaData
							resultEntry.MetaData = &metaData
							go updateMetaDataEc2(&resultWg, &resultData, &metaData, cloudAcc, regionName)
							resultWg.Add(1)
							//TODO get recommendations

						}

						rList = append(rList, resultEntry)
					}

					IList = rList
				} else if resourceName == "ebs" {
					var rList []aws_model.AwsBlockVolumeResult
					var policyresultList []aws_model.AwsBlockVolumePolicyResultData
					json.Unmarshal([]byte(resourceList), &policyresultList)
					for _, elem := range policyresultList {
						var resultData aws_model.AwsBlockVolumeResultData
						resultData.AvailabilityZone = elem.AvailabilityZone
						resultData.Encrypted = elem.Encrypted
						resultData.Region = elem.Region
						resultData.SnapshotId = elem.SnapshotId
						resultData.State = elem.State
						resultData.VolumeId = elem.VolumeId
						resultData.VolumeType = elem.VolumeType

						if len(elem.Attachments) == 0 {
							resultData.Attachments = false
						} else {
							resultData.Attachments = true
						}
						var resultEntry aws_model.AwsBlockVolumeResult
						resultEntry.ResultData = resultData
						resultEntry.MetaData = nil
						if policy.PolicyType == "Default" {
							var metaData model.ResultMetaData
							resultEntry.MetaData = &metaData
							go updateMetaDataEbs(&resultWg, &resultData, &metaData, cloudAcc, regionName)
							resultWg.Add(1)
							//TODO get recommendations
						}

						rList = append(rList, resultEntry)
					}
					IList = rList
				} else if resourceName == "elastic-ip" {
					var rList []aws_model.AwsElasticIPResult
					var policyresultList []aws_model.AwsElasticIPResultData
					json.Unmarshal([]byte(resourceList), &policyresultList)
					for _, elem := range policyresultList {
						var resultEntry aws_model.AwsElasticIPResult
						resultEntry.ResultData = elem
						resultEntry.MetaData = nil
						if policy.PolicyType == "Default" {
							var metaData model.ResultMetaData
							resultEntry.MetaData = &metaData
							go updateMetaDataEip(&resultWg, &elem, &metaData, cloudAcc, policy.Recommendation)
							resultWg.Add(1)
						}

						rList = append(rList, resultEntry)
					}
					IList = rList
				} else if resourceName == "ebs-snapshot" {
					var rList []aws_model.AwsSnapshotResult
					var policyresultList []aws_model.AwsSnapshotResultData
					json.Unmarshal([]byte(resourceList), &policyresultList)
					for _, elem := range policyresultList {
						var resultEntry aws_model.AwsSnapshotResult
						resultEntry.ResultData = elem
						resultEntry.MetaData = nil
						if policy.PolicyType == "Default" {
							var metaData model.ResultMetaData
							resultEntry.MetaData = &metaData
							go updateMetaDataAwsSnapshot(&resultWg, &elem, &metaData, cloudAcc, policy.Recommendation)
							resultWg.Add(1)
						}

						rList = append(rList, resultEntry)
					}
					//out, _ = json.Marshal(rList)
					IList = rList

				} else {
					fmt.Println("Unknown resource type ", resourceName)
					IList = nil
				}

				//regionResult.Result = string(out)
				regionResult.Result = IList
				regionResult.Region = regionName
				resultList = append(resultList, *regionResult)

			}
			//wait for threads
			resultWg.Wait()
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, resourceName, "SUCCESS", policyRunTime, resultList, true)
		}
		rchan <- isPolicyRunFailed
		return
	}
	rchan <- false
}

func updateMetaDataEc2(resultWg *sync.WaitGroup, result *aws_model.AwsInstanceResultData, resultMetaData *model.ResultMetaData, cloudAcc model.CloudAccountData, regionName string) {
	defer resultWg.Done()

	estimate, err := aws_cost_estimator.GetAWSRecommendationForEC2Instance(cloudAcc.AwsCredentials.AccessKeyID, cloudAcc.AwsCredentials.SecretAccessKey, regionName, cloudAcc.AwsCredentials.AccountID, result.InstanceId)

	if err != nil || estimate == nil {

		fmt.Printf("Failed to get recommendation %+v", err)
		product := aws_model.ProductAttributesInstance{
			InstanceType:    result.InstanceType,
			RegionCode:      regionName,
			OperatingSystem: "Linux",
		}
		cost, err := aws_cost_estimator.GetComputeInstanceCost(aws_model.ProductInfo[aws_model.ProductAttributesInstance]{
			Attributes: product})

		if err != nil {
			resultMetaData.Cost = "unknown"
			resultMetaData.Recommendations = nil
			return
		}
		resultMetaData.Cost = getMonthlyPrice(cost.MinPrice, cost.Currency, cost.Unit)
		resultMetaData.Recommendations = nil
		return
	} else {
		resultMetaData.Cost = getMonthlyPrice(estimate.CurrentCost.MinPrice, estimate.CurrentCost.Currency, estimate.CurrentCost.Unit)
		var recommendationList []model.ResultRecommendation
		for _, elem := range estimate.RecommendationItems {
			var recommendation model.ResultRecommendation
			recommendation.Price = getMonthlyPrice(elem.Cost.MinPrice, elem.Cost.Currency, elem.Cost.Unit)
			recommendation.Recommendation = elem.Resource.InstanceType
			recommendation.EstimatedCostSavings = elem.EstimatedCostSavings
			recommendation.EstimatedMonthlySavings = elem.EstimatedMonthlySavings
			recommendationList = append(recommendationList, recommendation)
		}
		resultMetaData.Recommendations = recommendationList
	}
}

func updateMetaDataEbs(resultWg *sync.WaitGroup, result *aws_model.AwsBlockVolumeResultData, resultMetaData *model.ResultMetaData, cloudAcc model.CloudAccountData, regionName string) {
	defer resultWg.Done()
	estimate, err := aws_cost_estimator.GetAWSRecommendationForEBSVolume(cloudAcc.AwsCredentials.AccessKeyID, cloudAcc.AwsCredentials.SecretAccessKey, regionName, cloudAcc.AwsCredentials.AccountID, result.VolumeId)
	//TODO
	if err != nil || estimate == nil {

		fmt.Printf("Failed to get recommendation %+v", err)
		product := aws_model.ProductAttributesEBS{
			VolumeType: result.VolumeType,
			RegionCode: regionName,
		}
		cost, err := aws_cost_estimator.GetEbsCost(aws_model.ProductInfo[aws_model.ProductAttributesEBS]{
			Attributes: product, ProductFamily: "Storage"})

		if err != nil {
			resultMetaData.Cost = "unknown"
			resultMetaData.Recommendations = nil
			return
		}
		resultMetaData.Cost = getMonthlyPrice(cost.MinPrice, cost.Currency, cost.Unit)
		resultMetaData.Recommendations = nil
		return

	} else {
		resultMetaData.Cost = getMonthlyPrice(estimate.CurrentCost.MinPrice, estimate.CurrentCost.Currency, estimate.CurrentCost.Unit)
		var recommendationList []model.ResultRecommendation
		for _, elem := range estimate.RecommendationItems {
			var recommendation model.ResultRecommendation
			recommendation.Price = getMonthlyPrice(elem.Cost.MinPrice, elem.Cost.Currency, elem.Cost.Unit)
			recommendation.Recommendation = elem.Resource.VolumeType
			recommendation.EstimatedCostSavings = elem.EstimatedCostSavings
			recommendation.EstimatedMonthlySavings = elem.EstimatedMonthlySavings
			recommendationList = append(recommendationList, recommendation)
		}
		resultMetaData.Recommendations = recommendationList
	}
}

func updateMetaDataEip(resultWg *sync.WaitGroup, result *aws_model.AwsElasticIPResultData, resultMetaData *model.ResultMetaData, cloudAcc model.CloudAccountData, defaultRecommendation string) {
	defer resultWg.Done()
	//estimate, err := aws_cost_estimator.GetAWSRecommendationForEBSVolume(cloudAcc.AwsCredentials.AccessKeyID, cloudAcc.AwsCredentials.SecretAccessKey, regionName, cloudAcc.AwsCredentials.AccountID, result.VolumeId)
	//TODO
	montlyCost := getMonthlyPrice(0.1, "USD", "Monthly")
	resultMetaData.Cost = montlyCost
	var recommendationList []model.ResultRecommendation

	var recommendation model.ResultRecommendation
	recommendation.Price = montlyCost
	recommendation.Recommendation = defaultRecommendation
	recommendation.EstimatedCostSavings = montlyCost
	recommendation.EstimatedMonthlySavings = "100%"

	recommendationList = append(recommendationList, recommendation)
	resultMetaData.Recommendations = recommendationList

}

func updateMetaDataAwsSnapshot(resultWg *sync.WaitGroup, result *aws_model.AwsSnapshotResultData, resultMetaData *model.ResultMetaData, cloudAcc model.CloudAccountData, defaultRecommendation string) {
	defer resultWg.Done()
	//estimate, err := aws_cost_estimator.GetAWSRecommendationForEBSVolume(cloudAcc.AwsCredentials.AccessKeyID, cloudAcc.AwsCredentials.SecretAccessKey, regionName, cloudAcc.AwsCredentials.AccountID, result.VolumeId)
	//TODO
	montlyCost := getMonthlyPrice(0.1, "USD", "Monthly")
	resultMetaData.Cost = montlyCost
	var recommendationList []model.ResultRecommendation

	var recommendation model.ResultRecommendation
	recommendation.Price = montlyCost
	recommendation.Recommendation = defaultRecommendation
	recommendation.EstimatedCostSavings = montlyCost
	recommendation.EstimatedMonthlySavings = "100%"

	recommendationList = append(recommendationList, recommendation)
	resultMetaData.Recommendations = recommendationList

}

func getMonthlyPrice(price float64, currency string, unit string) string {
	if strings.Contains(strings.ToLower(unit), "hrs") {
		return fmt.Sprintf("%f %s/Month", price*30, currency)
	}
	return fmt.Sprintf("%f %s/Month", price, currency)
}

func updatePolicyRunResult(pipeLineID string, policyID string, resourceName string, runStatus string, runtime int64, regionWiseResult []model.RegionResult, isSuccess bool) {
	opr := storage.GetDefaultDBOperators()
	query := fmt.Sprintf(`{"policyid": "%s","pipelineid": "%s"}`, policyID, pipeLineID)
	results, _ := opr.PolicyOperator.GetPolicyResultDetails(query)
	if len(results) == 0 {
		var policyRunresult model.PolicyResult
		policyRunresult.PipelIneID = pipeLineID
		policyRunresult.PolicyID = policyID
		policyRunresult.Resource = resourceName
		policyRunresult.LastRunStatus = runStatus
		policyRunresult.Resultlist = regionWiseResult
		policyRunresult.LastRunTime = runtime
		json.Unmarshal([]byte(getDisplayDefinition(resourceName)), &policyRunresult.DisplayDefinition)
		opr.PolicyOperator.AddPolicyResult(policyRunresult)

	} else {
		result := results[0]
		if isSuccess {
			result.Resultlist = regionWiseResult
		}
		result.LastRunStatus = runStatus
		result.LastRunTime = runtime
		result.Resource = resourceName
		result.PolicyID = policyID
		result.PipelIneID = pipeLineID
		json.Unmarshal([]byte(getDisplayDefinition(resourceName)), &result.DisplayDefinition)
		opr.PolicyOperator.UpdatePolicyResult(result)
	}
}

func getDisplayDefinition(resourceName string) string {
	if resourceName == "ec2" {
		return model.GetAWSInstanceDisplayDefinition()
	} else if resourceName == "ebs" {
		return model.GetAWSVolumeDisplayDefinition()
	} else if resourceName == "elastic-ip" {
		return model.GetAWSEIPDisplayDefinition()
	} else if resourceName == "ebs-snapshot" {
		return model.GetAWSSnapshotDisplayDefinition()
	}
	return ""
}
