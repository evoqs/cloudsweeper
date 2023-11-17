package utils

import (
	"bytes"
	"cloudsweep/config"
	"context"
	"encoding/json"
	"os"

	gl "cloud.google.com/go/ai/generativelanguage/apiv1beta2"
	pb "cloud.google.com/go/ai/generativelanguage/apiv1beta2/generativelanguagepb"
	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
)

func QueryOpenAiRest(query string) (string, error) {
	apiKey := config.GetConfig().OpenAI.ApiKey
	apiUrl := config.GetConfig().OpenAI.Url

	// Build the request body with the initial message for GPT-3.5 Turbo
	requestBody := []byte(`{
        "model": "gpt-4",
        "messages": [ ` + query + `],
        "temperature": 0.2
    }`)

	// Make a POST request to the API using the CallAPI function from the utils package
	postResponse, err := CallAPI("POST", apiUrl, apiKey, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	// Unmarshal the response JSON
	//fmt.Printf("postResponse.Body:%v", string(postResponse.Body))
	var responseMap map[string]interface{}
	if err := json.Unmarshal(postResponse.Body, &responseMap); err != nil {
		return string(postResponse.Body), err
	}

	// Extract the "content" from the response
	if choices, ok := responseMap["choices"].([]interface{}); ok {
		if len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if contentStr, ok := message["content"].(string); ok {
						return contentStr, nil
					}
				}
			}
		}
	}
	return string(postResponse.Body), nil
}

func QueryOpenAi(query string) (string, error) {
	client := openai.NewClient(config.GetConfig().OpenAI.ApiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			Temperature: 0.2,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func QueryMakerSuite(query string) (string, error) {
	ctx := context.Background()
	client, err := gl.NewTextRESTClient(ctx, option.WithAPIKey(os.Getenv("PALM_KEY")))
	if err != nil {
		panic(err)
	}
	defer client.Close()
	temperature := float32(0.2)
	req := &pb.GenerateTextRequest{
		Model:       "models/text-bison-001",
		Temperature: &temperature,
		Prompt: &pb.TextPrompt{
			Text: query,
		},
	}
	resp, err := client.GenerateText(ctx, req)
	if err != nil {
		panic(err)
	}
	//fmt.Println(resp.Candidates[0].Output)
	return resp.Candidates[0].Output, nil
}
