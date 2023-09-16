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
	//pipelineQueue <- pipeLine
	go runPipeline(pipeLine)
	return 200, nil
}

func runPipeline(pipeLine model.PipeLine) {
	logger.NewDefaultLogger().Info("Running the pipeline: " + pipeLine.PipeLineID.String())
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
			isPolicyRunFailed = true
			continue
		}

		if len(policyList) == 0 {
			fmt.Println("Policy not found")
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

		//Step 1 TODO
		//Step 2 create runfolder
		RunFolder := fmt.Sprintf("/tmp/%s.%s", policyid, strconv.Itoa(rand.Intn(100000)))
		os.Mkdir(RunFolder, os.ModePerm)

		//Step 3
		policyFile := fmt.Sprintf("%s/%s", RunFolder, "policy.yml")
		err = policy_converter.ConvertJsonToYamlAndWriteToFile(policyJson, policyFile)
		if err != nil {
			fmt.Println("Failed to convert json policy to yaml, for policy id ", policy.PolicyID)
			isPolicyRunFailed = true
			continue
		}

		//Get creds
		cloudAccList, err := opr.AccountOperator.GetCloudAccount(policy.CloudAccountID)
		if err != nil || len(cloudAccList) < 1 {
			fmt.Println("Failed to get cloundaccount details for policy id ", policy.PolicyID, policy.CloudAccountID)
			isPolicyRunFailed = true
			continue
		}

		cloudAcc := cloudAccList[0]
		if cloudAcc.AccountType == model.AWS {
			var envvars []string
			envvars = append(envvars, fmt.Sprintf("AWS_DEFAULT_REGION=%s", cloudAcc.AwsCredentials.Region))
			envvars = append(envvars, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", cloudAcc.AwsCredentials.AccessKeyID))
			envvars = append(envvars, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", cloudAcc.AwsCredentials.SecretAccessKey))

			activatePath := utils.GetConfig().Custodian.C7nAwsInstall
			c := make(chan string, 1)
			go utils.RunCustodianPolicy(envvars, RunFolder, policyFile, activatePath, c)
			runres, ok := <-c
			close(c)
			if !ok {
				fmt.Println("Failed to read from channel")
				isPolicyRunFailed = true
				continue
			}

			if strings.Contains(runres, "ERROR") {
				fmt.Println("policy run failed with result", runres)
				isPolicyRunFailed = true
			} else {
				fmt.Println("policy run successful with result", runres)
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
					var policyRunresult model.PolicyResult
					policyRunresult.PolicyID = policy.PolicyID.Hex()
					policyRunresult.Result = resourceList
					policyRunresult.Resource = resourceName

					query := fmt.Sprintf(`{"policyid": "%s"}`, policyRunresult.PolicyID)
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
					}
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
