package policy_converter

import (
	"os"
	"strings"
	"sync"

	"github.com/itchyny/json2yaml"
)

var (
	mutex sync.Mutex
)

// Converts the json file and returns the content into an file
//
// Converts the json file and returns the content into an file.
// It returns error when an error is encountered.
func convertJsonToYaml(json string) (string, error) {
	var yaml strings.Builder
	if err := json2yaml.Convert(&yaml, strings.NewReader(json)); err != nil {
		return "", err
	}
	return yaml.String(), nil
}

// Converts the json file and writes the content into an file
//
// Converts the json file and writes the content into an file.
// It returns error when an error is encountered.
func convertJsonToYamlAndWriteToFile(json string, path string) error {
	var yaml strings.Builder
	if err := json2yaml.Convert(&yaml, strings.NewReader(json)); err != nil {
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	if file, err := os.Create(path); err != nil {
		return err
	} else {
		defer file.Close()
		if _, err := file.Write([]byte(yaml.String())); err != nil {
			return err
		}
		return nil
	}
}
