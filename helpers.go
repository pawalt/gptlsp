package main

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// createChatCompletion is a function that creates a chat completion request
// and sends it to the OpenAI client.
func createChatCompletion(client *openai.Client, model string, messages []openai.ChatCompletionMessage, functionMetadata []*openai.FunctionDefine) (openai.ChatCompletionResponse, error) {
	// Create a ChatCompletionRequest object with the specified model, messages, temperature, and functions.
	r := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.7,
		Functions:   functionMetadata,
	}

	// Send the chat completion request to the OpenAI client and return the response.
	return client.CreateChatCompletion(context.Background(), r)
}