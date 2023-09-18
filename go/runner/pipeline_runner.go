package runner

import (
	logger "cloudsweep/logging"
	"cloudsweep/model"
	"cloudsweep/policy_converter"
	"cloudsweep/storage"
	"cloudsweep/utils"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
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
	policy := policies[0]
	accounts, err := opr.AccountOperator.GetCloudAccount(policy.CloudAccountID)
	if err != nil {
		return 500, err
	}
	if len(accounts) == 0 {
		return 404, fmt.Errorf("Account not found with id %s", policy.CloudAccountID)
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
	logger.NewDefaultLogger().Info("Running the pipeline: " + pipeLine.PipeLineID.String())
	//Step 1
	if pipeLine.RunStatus == model.RUNNING {
		fmt.Println("Run is already in progress")
		return
	}
	pipeLine.RunStatus = model.RUNNING
	opr := storage.GetDefaultDBOperators()
	upcount, err := opr.PipeLineOperator.UpdatePipeLine(pipeLine)
	if upcount != 1 {
		fmt.Println("Failed to update pipline run status,", err)
		return
	}

	policyids := pipeLine.Policies
	isPolicyRunFailed := false
	for _, policyid := range policyids {

		policyList, err := opr.PolicyOperator.GetPolicyDetails(policyid)
		if err != nil {
			fmt.Println("Failed to get policy")
			updatePolicyRunResult(policyid, "", "Internal DB Error", nil, false)
			isPolicyRunFailed = true
			continue
		}

		if len(policyList) == 0 {
			fmt.Println("Policy not found")
			updatePolicyRunResult(policyid, "", "Policy Definition missing", nil, false)
			isPolicyRunFailed = true
			continue
		}
		//fmt.Println("Length", len(policyList))
		policy := policyList[0]
		policyJson := policy.PolicyDefinition

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

		//Step 3
		policyFile := fmt.Sprintf("%s/%s", RunFolder, "policy.yml")
		err = policy_converter.ConvertJsonToYamlAndWriteToFile(policyJson, policyFile)
		if err != nil {
			fmt.Println("Failed to convert json policy to yaml, for policy id ", policy.PolicyID)
			updatePolicyRunResult(policyid, "", "Invalid policy definition", nil, false)
			isPolicyRunFailed = true
			continue
		}

		//Get creds
		cloudAccList, err := opr.AccountOperator.GetCloudAccount(policy.CloudAccountID)
		if err != nil || len(cloudAccList) < 1 {
			fmt.Println("Failed to get cloundaccount details for policy id ", policy.PolicyID, policy.CloudAccountID)
			updatePolicyRunResult(policyid, "", "Missing Cloud Account definition for policy", nil, false)
			isPolicyRunFailed = true
			continue
		}

		cloudAcc := cloudAccList[0]
		//validating cloud authentication

		if cloudAcc.AccountType == model.AWS {
			if !utils.ValidateAwsCredentials(cloudAcc.AwsCredentials.AccessKeyID, cloudAcc.AwsCredentials.SecretAccessKey) {
				updatePolicyRunResult(policyid, "", "Authentication Failed", nil, false)
				isPolicyRunFailed = true
				continue
			}
			var envvars []string
			envvars = append(envvars, fmt.Sprintf("AWS_DEFAULT_REGION=%s", policy.ExecutionRegions[0]))
			envvars = append(envvars, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", cloudAcc.AwsCredentials.AccessKeyID))
			envvars = append(envvars, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", cloudAcc.AwsCredentials.SecretAccessKey))

			activatePath := utils.GetConfig().Custodian.C7nAwsInstall
			c := make(chan string, 1)
			go utils.RunCustodianPolicy(envvars, RunFolder, policyFile, activatePath, c)
			runres, ok := <-c
			close(c)
			if !ok {
				fmt.Println("Failed to read from channel")
				updatePolicyRunResult(policyid, "", "Internal Error", nil, false)
				isPolicyRunFailed = true
				continue
			}

			if strings.Contains(strings.ToUpper(runres), "ERROR") {
				fmt.Println("policy run failed with result", runres)
				updatePolicyRunResult(policyid, "", "Internal Error", nil, false)
				isPolicyRunFailed = true
			} else {
				fmt.Println("policy run successful with result", runres)
				var resultList = make([]model.RegionResult, 0)
				fmt.Println(resultList)
				regionResult := new(model.RegionResult)

				//get policy name and resource
				policyName, _ := utils.GetFirstMatchingGroup(runres, "policy:.*?policy:(.*?)\\sresource:")
				if policyName != "" {
					resourceFile := fmt.Sprintf("%s/%s/%s", RunFolder, policyName, "resources.json")
					resourceList, err := utils.ReadFile(resourceFile)
					replacer := strings.NewReplacer("\r", "", "\n", "")
					resourceList = replacer.Replace(string(resourceList))
					//var jsonMap map[string]interface{}
					//json.Unmarshal([]byte(resourceList), &jsonMap)
					if err != nil {
						fmt.Println("Failed to read policy result from result", resourceFile)
						isPolicyRunFailed = true
						continue
					} else {
						fmt.Println("*********", resourceList)
					}
					resourceName, err := utils.GetFirstMatchingGroup(runres, "resource:(.*?)\\s")

					regionResult.Result = resourceList
					regionResult.Region = policy.ExecutionRegions[0]
					resultList = append(resultList, *regionResult)
					updatePolicyRunResult(policyid, resourceName, "SUCCESS", resultList, true)
					/*policyRunresult := new(model.PolicyResult)

					policyRunresult.PolicyID = policy.PolicyID.Hex()
					regionResult.Result = resourceList
					policyRunresult.Resource = resourceName
					regionResult.Region = policy.ExecutionRegions[0]*/

					/*	query := fmt.Sprintf(`{"policyid": "%s"}`, policyRunresult.PolicyID)
						results, err := opr.PolicyOperator.GetPolicyResultDetails(query)
						if err != nil {
							fmt.Println("Failed to read policy result from DB", policyRunresult.PolicyID)
							isPolicyRunFailed = true
							continue
						}
						if len(results) == 0 {
							opr.PolicyOperator.AddPolicyResult(policyRunresult)
						} else {
							result := results[0]
							result.Result = resourceList
							opr.PolicyOperator.UpdatePolicyResult(result)
						}*/
				}
			}
		}

	}
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

func updatePolicyRunResult(policyID string, resourceName string, runStatus string, regionWiseResult []model.RegionResult, isSuccess bool) {
	opr := storage.GetDefaultDBOperators()
	query := fmt.Sprintf(`{"policyid": "%s"}`, policyID)
	results, _ := opr.PolicyOperator.GetPolicyResultDetails(query)
	if len(results) == 0 {
		var policyRunresult model.PolicyResult
		policyRunresult.PolicyID = policyID
		policyRunresult.Resource = resourceName
		policyRunresult.LastRunStatus = runStatus
		policyRunresult.Resultlist = regionWiseResult
		opr.PolicyOperator.AddPolicyResult(policyRunresult)

	} else {
		result := results[0]
		if isSuccess {
			result.Resultlist = regionWiseResult
		}
		result.LastRunStatus = runStatus
		result.Resource = resourceName
		result.PolicyID = policyID
		opr.PolicyOperator.UpdatePolicyResult(result)
	}
}
