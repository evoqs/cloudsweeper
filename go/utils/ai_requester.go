package utils

import (
	"bytes"
)

func QueryOpenAI(query string) (*APIResponse, error) {
	apiKey := GetConfig().OpenAI.ApiKey
	apiUrl := GetConfig().OpenAI.Url

	// Build the request body with the initial message for GPT-3.5 Turbo
	requestBody := []byte(`{
        "model": "gpt-4",
        "messages": [ ` + query + `],
        "temperature": 0.2
    }`)

	// Make a POST request to the API using the CallAPI function from the utils package
	postResponse, err := CallAPI("POST", apiUrl, apiKey, bytes.NewBuffer(requestBody))
	if err != nil {
		return &APIResponse{}, err
	}
	return postResponse, nil
}
