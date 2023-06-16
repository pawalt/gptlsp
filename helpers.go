package main

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

func createChatCompletion(client *openai.Client, model string, messages []openai.ChatCompletionMessage, functionMetadata []*openai.FunctionDefine) (openai.ChatCompletionResponse, error) {
	r := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.7,
		Functions:   functionMetadata,
	}
	return client.CreateChatCompletion(context.Background(), r)
}
