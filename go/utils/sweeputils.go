package utils

import (
	"bufio"
	"cloudsweep/config"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func GetDBUrl(c *config.Config) (string, error) {
	var mongoDBUrl string
	mongoDBUrl = "mongodb://"
	if c.Database.Username != "" {
		if c.Database.Password != "" {
			mongoDBUrl = mongoDBUrl + strings.TrimSpace(c.Database.Username) + ":" + url.QueryEscape(strings.TrimSpace(c.Database.Password)) + "@"
		} else {
			return "", errors.New("Empty Password for mongodb")
		}

	}

	if c.Database.Host != "" {
		mongoDBUrl = mongoDBUrl + strings.TrimSpace(c.Database.Host)
	} else {
		return "", errors.New("Empty Hostname for mongodb")
	}

	if c.Database.Port != "" {
		mongoDBUrl = mongoDBUrl + ":" + strings.TrimSpace(c.Database.Port)
	}

	if c.Database.Name != "" {
		mongoDBUrl = mongoDBUrl + "/" + strings.TrimSpace(c.Database.Name)
	}

	return mongoDBUrl, nil
}

func RunPolicy() string {
	//cmd := "cat /proc/cpuinfo | egrep '^model name' | uniq | awk '{print substr($0, index($0,$4))}'"
	cmd := "source /home/vboxuser/custodian/bin/activate; custodian run --dryrun -s /tmp /home/vboxuser/c7npolicies/policy1.yml"
	out := exec.Command("bash", "-c", cmd)
	out.Env = append(out.Environ(), "AWS_DEFAULT_REGION=ap-southeast-2")
	out.Env = append(out.Environ(), "AWS_ACCESS_KEY_ID=AKIA4T2VWH7A6GQYCS7Z")
	out.Env = append(out.Environ(), "AWS_SECRET_ACCESS_KEY=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")
	stdout, err := out.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to execute command: %s \n %s \n %s", cmd, err, stdout)
	}
	return string(stdout)
}

func RunCustodianPolicy(envvars []string, runfolder string, policyfile string, activation string, regionlist string, runchan chan string) {
	//cmd := "cat /proc/cpuinfo | egrep '^model name' | uniq | awk '{print substr($0, index($0,$4))}'"
	//cmd := "source /home/vboxuser/custodian/bin/activate; custodian run --dryrun -s /tmp /home/vboxuser/c7npolicies/policy1.yml"

	cmd := fmt.Sprintf("source %s; custodian run --dryrun -s %s %s %s", activation, runfolder, regionlist, policyfile)
	out := exec.Command("bash", "-c", cmd)
	for _, envvar := range envvars {
		out.Env = append(out.Environ(), envvar)
	}

	stdout, err := out.CombinedOutput()
	if err != nil {
		runchan <- fmt.Sprintf("Failed to execute command: %s \n %s \n %s", cmd, err, stdout)
	}
	runchan <- string(stdout)
}

func ValidateAwsCredentials(accesskey string, accesssecret string) bool {
	/*var envvars []string
	envvars = append(envvars, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", accesskey))
	envvars = append(envvars, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", accesssecret))*/
	cmd := "aws sts get-caller-identity"
	out := exec.Command("bash", "-c", cmd)
	out.Env = append(out.Environ(), fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", accesskey))
	out.Env = append(out.Environ(), fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", accesssecret))
	stdout, err := out.CombinedOutput()
	if err != nil {
		fmt.Printf(fmt.Sprintf("Failed to execute command: %s \n %s \n %s", cmd, err, stdout))
		return false
	}

	fmt.Println("Exit code ===========> ", out.ProcessState.ExitCode())
	fmt.Println(string(stdout))
	if strings.Contains(string(stdout), "error") {
		return false
	} else {
		return true
	}

}

// Fetches the regions subscribed by an account
func GetAwsSubscribedRegions(accesskey string, accesssecret string) bool {
	/*var envvars []string
	envvars = append(envvars, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", accesskey))
	envvars = append(envvars, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", accesssecret))*/
	cmd := "aws ec2 describe-availability-zones | jq '.AvailabilityZones[].RegionName' | sort | uniq"
	out := exec.Command("bash", "-c", cmd)
	out.Env = append(out.Environ(), fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", accesskey))
	out.Env = append(out.Environ(), fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", accesssecret))
	stdout, err := out.CombinedOutput()
	if err != nil {
		fmt.Printf(fmt.Sprintf("Failed to execute command: %s \n %s \n %s", cmd, err, stdout))
		return false
	}

	fmt.Println("Exit code ===========> ", out.ProcessState.ExitCode())
	fmt.Println(string(stdout))
	if strings.Contains(string(stdout), "error") {
		return false
	} else {
		return true
	}

}

func GetFirstMatchingGroup(sentence string, regex string) (string, error) {

	rgx, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}

	rs := rgx.FindStringSubmatch(sentence)
	if len(rs) >= 2 {
		return rs[1], nil
	}
	return "", errors.New("No matching group Found")

}

func ReadFile(fileName string) (string, error) {

	fileContent, err := ioutil.ReadFile(fileName)
	return string(fileContent), err

}

func ConstructRegionList(listOfregions *[]string) string {

	var regionCmd string
	for _, region := range *listOfregions {
		regionCmd = fmt.Sprintf("%s --region %s", regionCmd, region)
	}
	return strings.TrimSpace(regionCmd)
}

func GetFolderList(path string) []string {
	entries, err := os.ReadDir("/home/vboxuser/c7npolicies")
	folderList := []string{}
	if err != nil {
		//logger.NewDefaultLogger().Logger.Error("Unable to read folder ", path)
	}
	for _, e := range entries {
		if e.IsDir() {
			folderList = append(folderList, e.Name())
		}
	}

	return folderList
}

func GetResourceName(yamlFile string) string {
	file, err := os.Open(yamlFile)
	if err != nil {
		//logger.NewDefaultLogger().Logger.Error(err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix((line), "resource:") {
			return strings.Split(line, " ")[1]
		}
	}

	return ""
}

func ExtractJsonFromString(input string) ([]string, error) {
	jsonRegex := regexp.MustCompile(`\{.*\}`)

	// Find the first match
	matches := jsonRegex.FindStringSubmatch(input)
	if len(matches) == 0 {
		return []string{}, fmt.Errorf("JSON not found in the input string")
	}
	return matches, nil
}
