package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/sashabaranov/go-openai"
)

type AnalyzeRequest struct {
	Instruction string `json:"instruction,omitempty"`
}

func Analyze(raw string, client *openai.Client) map[string]string {
	var req AnalyzeRequest
	_ = json.Unmarshal([]byte(raw), &req)
	if req.Instruction == "" {
		return map[string]string{
			"error": "instruction must be specified",
		}
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "Use the functions available to you to answer the user's queries.",
		},
		{
			Role:    "user",
			Content: req.Instruction,
		},
	}

	done := false
	for !done {
		// model := openai.GPT3Dot5Turbo0613
		model := openai.GPT40613
		r := openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: 0.7,
			Functions: []*openai.FunctionDefine{
				getTimeMetadata,
				listFilesMetadata,
				readFilesMetadata,
			},
		}

		completions, err := client.CreateChatCompletion(context.Background(), r)
		if err != nil {
			log.Panicf("could not create completions %v", err)
		}

		if len(completions.Choices) != 1 {
			log.Panicf("somehow not length 1 choices %v", completions.Choices)
		}

		choice := completions.Choices[0]

		messages = append(messages, choice.Message)

		switch choice.FinishReason {
		case "function_call":
			messages = append(messages, executeFunction(
				client,
				choice.Message.FunctionCall.Name,
				choice.Message.FunctionCall.Arguments,
			))
		default:
			done = true
		}
	}

	return map[string]string{
		"analysis": messages[len(messages)-1].Content,
	}
}

var analyzeMetadata = &openai.FunctionDefine{
	Name:        "analyze",
	Description: "Analyze files in a project",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"instruction": {
				Type:        openai.JSONSchemaTypeString,
				Description: `Natural language description of what analysis to perform`,
			},
		},
		Required: []string{"instruction"},
	},
}
