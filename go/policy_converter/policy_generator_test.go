package policy_generator

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestJsonToYamlPrint(t *testing.T) {
	json := `{
		"policies": [
		  {
			"name": "ec2-require-imdsv2",
			"resource": "ec2",
			"description": "Finds all instances with optional HttpTokens and change the policy to Requied.\n",
			"filters": [
			  {
				"MetadataOptions.HttpTokens": "optional"
			  }
			],
			"actions": [
			  {
				"type": "set-metadata-access",
				"tokens": "required"
			  }
			]
		  }
		]
	}`
    if got,_ := convertJsonToYaml(json); len(got) != 0 {
        t.Logf("Yaml %v", got)
    }
}

func TestJsonToYamlAndWriteToFilePrint(t *testing.T) {
	json := `{
		"policies": [
		  {
			"name": "ec2-require-imdsv2",
			"resource": "ec2",
			"description": "Finds all instances with optional HttpTokens and change the policy to Requied.\n",
			"filters": [
			  {
				"MetadataOptions.HttpTokens": "optional"
			  }
			],
			"actions": [
			  {
				"type": "set-metadata-access",
				"tokens": "required"
			  }
			]
		  }
		]
	}`
    if err := convertJsonToYamlAndWriteToFile(json, "/tmp/json2yaml"); err != nil {
        t.Fatalf("Failed to convert Json to Yaml and Write to File. %v", err)
    } else {
		t.Logf("Successfully converted json to yaml and wrote to the file.")
	}
}

func TestJsonToYamlSingle(t *testing.T) {
	json := `{
		"policies": [
		  {
			"name": "ec2-require-imdsv2",
			"resource": "ec2",
			"description": "Finds all instances with optional HttpTokens and change the policy to Requied.\n",
			"filters": [
			  {
				"MetadataOptions.HttpTokens": "optional"
			  }
			],
			"actions": [
			  {
				"type": "set-metadata-access",
				"tokens": "required"
			  }
			]
		  }
		]
	}`
    want := `policies:
	- name: ec2-require-imdsv2
	  resource: ec2
	  description: |
		Finds all instances with optional HttpTokens and change the policy to Requied.
	  filters:
		- MetadataOptions.HttpTokens: optional
	  actions:
		- type: set-metadata-access
		  tokens: required`
	
	replacements := map[string]string{
		" ":  "",
		"\n": "",	
		"\t": "",
	}
	want = replaceAll(want, replacements)
	if yaml, err := convertJsonToYaml(json); err != nil {
		t.Logf("Error %v", err)
	} else {
		if got := replaceAll(yaml, replacements); got != want {
			t.Errorf("jsonToYaml() = %q, want %q", got, want)
		}
	}
}

func replaceAll(input string, replacements map[string]string) string {
	output := input
	for oldStr, newStr := range replacements {
		output = strings.ReplaceAll(output, oldStr, newStr)
	}
	return output
}

func TestJsonToYamlRecurse(t *testing.T) {
	startTime := time.Now()
	json := `{
		"policies": [
		  {
			"name": "ec2-require-imdsv2",
			"resource": "ec2",
			"description": "Finds all instances with optional HttpTokens and change the policy to Requied.\n",
			"filters": [
			  {
				"MetadataOptions.HttpTokens": [
					{
						"optional": {
							"possibly-optional": [{
								"testone": "can-be-optional",
								"testtwo": "never-be-optional"
							}]
						}
					}
				]
			  }
			],
			"actions": [
			  {
				"type": "set-metadata-access",
				"tokens": "required"
			  }
			]
		  }
		]
	}`
	noOfIterations := 100000
	for iteration:=0; iteration < noOfIterations; iteration++ {
		if got,_ := convertJsonToYaml(json); len(got) == 0 {
			t.Errorf("jsonToYaml() = %q", got)
		}
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Printf("Took %v to convert %d jsons to yamls\n", duration, noOfIterations)
}

func TestJsonToYamlAndWriteToFileRecurse(t *testing.T) {
	startTime := time.Now()
	json := `{
		"policies": [
		  {
			"name": "ec2-require-imdsv2",
			"resource": "ec2",
			"description": "Finds all instances with optional HttpTokens and change the policy to Requied.\n",
			"filters": [
			  {
				"MetadataOptions.HttpTokens": [
					{
						"optional": {
							"possibly-optional": [{
								"testone": "can-be-optional",
								"testtwo": "never-be-optional"
							}]
						}
					}
				]
			  }
			],
			"actions": [
			  {
				"type": "set-metadata-access",
				"tokens": "required"
			  }
			]
		  }
		]
	}`
	noOfIterations := 100000
	for iteration:=0; iteration < noOfIterations; iteration++ {
		if err := convertJsonToYamlAndWriteToFile(json, "/tmp/json2yaml"); err != nil {
			t.Errorf("jsonToYaml() = %q", err)
		}
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Printf("Took %v to convert %d jsons to yamls\n", duration, noOfIterations)
}