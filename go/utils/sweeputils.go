package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
)

func GetDBUrl(c *Config) (string, error) {
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

func RunCustodianPolicy(envvars []string, runfolder string, policyfile string, activation string, runchan chan string) {
	//cmd := "cat /proc/cpuinfo | egrep '^model name' | uniq | awk '{print substr($0, index($0,$4))}'"
	//cmd := "source /home/vboxuser/custodian/bin/activate; custodian run --dryrun -s /tmp /home/vboxuser/c7npolicies/policy1.yml"

	cmd := fmt.Sprintf("source %s; custodian run --dryrun -s %s %s", activation, runfolder, policyfile)
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
