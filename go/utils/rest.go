package utils

import (
	"io"
	"net/http"
	"time"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	StatusCode int
	Body       []byte
}

// CallAPI makes a generic HTTP request to the API
func CallAPI(method, apiUrl, apiKey string, requestBody io.Reader) (*APIResponse, error) {
	timeout := 10 * time.Second
	// Create an HTTP request
	req, err := http.NewRequest(method, apiUrl, requestBody)
	if err != nil {
		return nil, err
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the HTTP request
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &APIResponse{
		StatusCode: resp.StatusCode,
		Body:       responseBody,
	}, nil
}
