package runner

import (
	"cloudsweep/config"
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
		return 409, errors.New("Cannot Run Pipeline in disabled state.")
	}
	//Validate the policy
	policyids := pipeLine.Policies
	if len(policyids) == 0 {
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
	logwriter.Infof("Running the pipeline: %s", pipeLine.PipeLineID.String())
	//Step 1
	if pipeLine.RunStatus == model.RUNNING {
		logwriter.Info("Run is already in progress")
		return
	}
	pipeLine.RunStatus = model.RUNNING
	opr := storage.GetDefaultDBOperators()
	upcount, err := opr.PipeLineOperator.UpdatePipeLine(pipeLine)
	if upcount != 1 {
		logwriter.Errorf("Failed to update pipline run status, %s", err.Error())
		return
	}

	//TODO Implement Parallel processing of polices
	policyids := pipeLine.Policies
	isPolicyRunFailed := false
	var wg sync.WaitGroup

	pipeLine.LastRunTime = time.Now().Unix()

	var cmap map[string]chan bool

	for _, policyid := range policyids {

		policyList, err := opr.PolicyOperator.GetPolicyDetails(policyid)
		logwriter.Infof("Running policy, %s", policyList[0].PolicyID.Hex())
		if err != nil {
			logwriter.Errorf("Failed to get policy, get DB operation failed, %s", err.Error())
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Internal DB Error", time.Now().Unix(), nil, false)
			isPolicyRunFailed = true
			continue
		}

		if len(policyList) == 0 {
			fmt.Println("Policy not found")
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Policy Definition missing", time.Now().Unix(), nil, false)
			isPolicyRunFailed = true
			continue
		}

		wg.Add(1)
		retunChan := make(chan bool, 1)
		cmap[policyList[0].PolicyID.Hex()] = retunChan
		go runPolicy(&wg, policyList[0], pipeLine, retunChan)

	}
	wg.Wait()

	for elem := range cmap {

		runResult := <-cmap[elem]
		logwriter.Infof("Completed policy run %s with result %t", elem, !runResult)

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
		rchan <- isPolicyRunFailed
		return
	}

	//Get creds
	cloudAccList, err := opr.AccountOperator.GetCloudAccount(pipeLine.CloudAccountID)
	if err != nil || len(cloudAccList) < 1 {
		logwriter.Errorf("Failed to get cloundaccount details for policy id %s %s , %s", policy.PolicyID, pipeLine.CloudAccountID, err.Error())
		updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Missing Cloud Account definition for policy", time.Now().Unix(), nil, false)
		isPolicyRunFailed = true
	}

	cloudAcc := cloudAccList[0]
	//validating cloud authentication

	if cloudAcc.AccountType == model.AWS {
		if !utils.ValidateAwsCredentials(cloudAcc.AwsCredentials.AccessKeyID, cloudAcc.AwsCredentials.SecretAccessKey) {
			logwriter.Errorf("AWS Authentication failed for cloud account %s, %s", cloudAcc.CloudAccountID.Hex(), err.Error())
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "AWS Authentication Failed, invalid accesskey/secret", time.Now().Unix(), nil, false)
			rchan <- isPolicyRunFailed
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
			rchan <- isPolicyRunFailed
			return

		}

		if strings.Contains(strings.ToUpper(runres), "ERROR") {
			fmt.Println("policy run failed with result", runres)
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, "", "Internal Error", policyRunTime, nil, false)
			isPolicyRunFailed = true
		} else {
			logger.NewDefaultLogger().Infof("policy run successful with result %s", runres)
			folderList := utils.GetFolderList(RunFolder)
			var resultList = make([]model.RegionResult, 0)
			resourceName := utils.GetResourceName(policyFile)

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

				var out []byte
				out = []byte("")

				if err != nil {
					logwriter.Errorf("Failed to read policy result from result %s, Error %s", resourceFile, err.Error())
					isPolicyRunFailed = true
					continue
				} else {
					logwriter.Infof("Successfully completed the policy run, %s", policyName)
				}

				//Converting into shorter json
				if resourceName == "ec2" {
					var rList []aws_model.AwsInstanceResult
					json.Unmarshal([]byte(resourceList), &rList)
					out, _ = json.Marshal(rList)
				} else if resourceName == "ebs" {
					var rList []aws_model.AwsBlockVolumeResult
					json.Unmarshal([]byte(resourceList), &rList)
					out, _ = json.Marshal(rList)
				} else {
					fmt.Println("Unknown resource type ", resourceName)
				}

				regionResult.Result = string(out)
				regionResult.Region = regionName
				resultList = append(resultList, *regionResult)

			}
			updatePolicyRunResult(pipeLine.PipeLineID.Hex(), policyid, resourceName, "SUCCESS", policyRunTime, resultList, true)
		}
		rchan <- isPolicyRunFailed
		return
	}
	rchan <- false
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
		opr.PolicyOperator.UpdatePolicyResult(result)
	}
}
