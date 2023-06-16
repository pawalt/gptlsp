package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/sashabaranov/go-openai"
)

type Request struct {
	Instruction string `json:"instruction,omitempty"`
}

func processRequest(raw string, client *openai.Client, systemMessage string, functionMetadata []*openai.FunctionDefine) map[string]string {
	var req Request
	_ = json.Unmarshal([]byte(raw), &req)
	if req.Instruction == "" {
		return map[string]string{
			"error": "instruction must be specified",
		}
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: systemMessage,
		},
		{
			Role:    "user",
			Content: req.Instruction,
		},
	}

	done := false
	for !done {
		model := openai.GPT40613
		r := openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: 0.7,
			Functions:   functionMetadata,
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
		"result": messages[len(messages)-1].Content,
	}
}

func Analyze(raw string, client *openai.Client) map[string]string {
	return processRequest(raw, client, "Use the functions available to you to answer the user's queries.", []*openai.FunctionDefine{
		getTimeMetadata,
		listFilesMetadata,
		readFilesMetadata,
	})
}

func Modify(raw string, client *openai.Client) map[string]string {
	return processRequest(raw, client, "Use the functions available to you to follow user's modification instructions.", []*openai.FunctionDefine{
		getTimeMetadata,
		listFilesMetadata,
		readFilesMetadata,
		writeFileMetadata,
	})
}

var analyzeMetadata = &openai.FunctionDefine{
	Name:        "analyze",
	Description: "Analyze files in a project. The analyzer can read and list files.",
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

var modifyFilesMetadata = &openai.FunctionDefine{
	Name:        "modify_files",
	Description: "Modify files in a project. The modifier can list, read, and write to files.",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"instruction": {
				Type:        openai.JSONSchemaTypeString,
				Description: `Natural language description of what modification to perform.\nThe instruction must contain specific instructions for what kind of modification to perform.`,
			},
		},
		Required: []string{"instruction"},
	},
}
