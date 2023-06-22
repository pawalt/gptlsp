package main

import (
	"encoding/json"
	"log"

	"github.com/sashabaranov/go-openai"
)

// Request represents the incoming request data
type Request struct {
	Instruction string `json:"instruction,omitempty"`
}

// processRequest processes the incoming request
func processRequest(
	raw string,
	client *openai.Client,
	systemMessage string,
	functionMetadata []*openai.FunctionDefine,
	history []openai.ChatCompletionMessage,
) map[string]string {
	var req Request

	// Unmarshal the raw request into the Request struct
	_ = json.Unmarshal([]byte(raw), &req)

	// If instruction is not specified, return an error
	if req.Instruction == "" {
		return map[string]string{
			"error": "instruction must be specified",
		}
	}

	// Append system message and user instruction to the history
	messages := append(history, []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: systemMessage,
		},
		{
			Role:    "user",
			Content: req.Instruction,
		},
	}...)

	done := false
	for !done {
		// Create chat completions
		completions, err := createChatCompletion(client, gpt4, messages, functionMetadata)
		if err != nil {
			log.Panicf("could not create completions %v", err)
		}

		// Check if the completions have only one choice
		if len(completions.Choices) != 1 {
			log.Panicf("somehow not length 1 choices %v", completions.Choices)
		}

		choice := completions.Choices[0]

		// Append the choice message to the messages
		messages = append(messages, choice.Message)

		switch choice.FinishReason {
		case "function_call":
			// Execute the function call
			messages = append(messages, executeFunction(
				client,
				choice.Message.FunctionCall.Name,
				choice.Message.FunctionCall.Arguments,
				messages,
				nil,
			))
		default:
			done = true
		}
	}

	return map[string]string{
		"result": messages[len(messages)-1].Content,
	}
}

// Analyze analyzes the request
func Analyze(raw string, client *openai.Client, history []openai.ChatCompletionMessage) map[string]string {
	return processRequest(raw, client, "Use the functions available to you to answer the user's queries.", []*openai.FunctionDefine{
		listFilesMetadata,
		readFilesMetadata,
	},
		history,
	)
}

// Modify modifies the request
func Modify(raw string, client *openai.Client, history []openai.ChatCompletionMessage) map[string]string {
	return processRequest(
		raw,
		client,
		"Use the functions available to you to modify files according to the user's instructions. Your are allowed to read and write files.",
		[]*openai.FunctionDefine{
			listFilesMetadata,
			readFilesMetadata,
			writeFileMetadata,
		},
		history,
	)
}

var analyzeMetadata = &openai.FunctionDefine{
	Name:        "analyze",
	Description: "Analyze files in a project. The analyzer can read and list files.",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"instruction": {
				Type:        openai.JSONSchemaTypeString,
				Description: `JSON-escaped natural language description of what analysis to perform`,
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
				Description: `JSON-escaped natural language description of what modification to perform.\nThe instruction must contain specific instructions for what kind of modification to perform.`,
			},
		},
		Required: []string{"instruction"},
	},
}