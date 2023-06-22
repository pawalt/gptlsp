package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type RefactorRequest struct {
	Instructions string   `json:"instructions,omitempty"`
	Paths        []string `json:"paths,omitempty"`
}

func Refactor(raw string, client *openai.Client) map[string]any {
	var req RefactorRequest
	_ = json.Unmarshal([]byte(raw), &req)
	if req.Instructions == "" {
		return map[string]any{
			"error": "instruction must be specified",
		}
	}

	for _, fp := range req.Paths {
		content, err := os.ReadFile(fp)
		e(err)

		messages := []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: fmt.Sprintf(`You are an expert refactoring bot. You should refactor
the provided Go code according to the following instructions, and you should only
respond with the refactored code with no decoration text.
 
Instructions:
%s`, req.Instructions),
			},
			{
				Role: "user",
				Content: fmt.Sprintf(`Code to refactor:
%s`, string(content)),
			},
		}

		completions, err := createChatCompletion(client, gpt3, messages, []*openai.FunctionDefine{})
		e(err)

		if len(completions.Choices) != 1 {
			log.Panicf("somehow not length 1 choices %v", completions.Choices)
		}

		choice := completions.Choices[0]

		err = os.WriteFile(fp, []byte(choice.Message.Content), 0755)
		e(err)
	}

	return map[string]any{
		"success": true,
	}
}

var refactorFilesMetadata = &openai.FunctionDefine{
	Name:        "refactor_files",
	Description: "Refactor files at the given paths",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"paths": {
				Type:        openai.JSONSchemaTypeArray,
				Description: `Relative paths to the files to refactor`,
				Items: &openai.JSONSchemaDefine{
					Type: openai.JSONSchemaTypeString,
				},
			},
			"instructions": {
				Type:        openai.JSONSchemaTypeString,
				Description: `JSON-escaped instructions on how to perform the refactor`,
			},
		},
		Required: []string{"paths", "instructions"},
	},
}

type SearchFilesRequest struct {
	Term string `json:"term,omitempty"`
}

func SearchFiles(raw string) map[string]any {
	var req SearchFilesRequest
	_ = json.Unmarshal([]byte(raw), &req)
	searchTerm := req.Term

	currDir, err := filepath.Abs(".")
	e(err)

	files, err := os.ReadDir(currDir)
	e(err)

	var output strings.Builder
	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".") || f.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(currDir, f.Name()))
		e(err)

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.Contains(line, searchTerm) {
				output.WriteString(fmt.Sprintf(`%s:%d:%s\n`, f.Name(), i, line))
			}
		}
	}

	return map[string]any{
		"matched_lines": output.String(),
	}
}

var searchFilesMetadata = &openai.FunctionDefine{
	Name:        "search_files",
	Description: "Search for a term in all the files in the current directory",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"term": {
				Type:        openai.JSONSchemaTypeString,
				Description: `Term to search for`,
			},
		},
		Required: []string{"term"},
	},
}
