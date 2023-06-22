package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// RefactorRequest represents the request body for the Refactor function.
type RefactorRequest struct {
	Instructions string   `json:"instructions,omitempty"` // Instructions for refactoring
	Paths        []string `json:"paths,omitempty"`        // Paths of files to refactor
}

// Refactor refactors the specified files according to the provided instructions.
// It returns a map with the result of the refactoring process.
func Refactor(raw string, client *openai.Client) map[string]interface{} {
	var req RefactorRequest
	_ = json.Unmarshal([]byte(raw), &req)
	if req.Instructions == "" {
		return map[string]interface{}{
			"error": "instruction must be specified",
		}
	}

	currDir, err := filepath.Abs(".")
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	paths, err := cleanPaths(req.Paths, currDir)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	for _, fp := range paths {
		content, err := os.ReadFile(fp)
		if err != nil {
			return map[string]interface{}{
				"error": err.Error(),
			}
		}

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
		if err != nil {
			return map[string]interface{}{
				"error": err.Error(),
			}
		}

		if len(completions.Choices) != 1 {
			log.Panicf("somehow not length 1 choices %v", completions.Choices)
		}

		choice := completions.Choices[0]

		err = os.WriteFile(fp, []byte(choice.Message.Content), 0755)
		if err != nil {
			return map[string]interface{}{
				"error": err.Error(),
			}
		}

		fmt.Printf("refactored %s\n", fp)
	}

	return map[string]interface{}{
		"success": true,
	}
}

// refactorFilesMetadata is the metadata for the refactor_files function.
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
				Description: `JSON-escaped natural language instructions on how to perform the refactor`,
			},
		},
		Required: []string{"paths", "instructions"},
	},
}

// SearchFilesRequest represents the request body for the SearchFiles function.
type SearchFilesRequest struct {
	Term string `json:"term,omitempty"` // Term to search for inside files
}

// SearchFiles searches for the specified term in all the files in the current directory.
// It returns a map with the matched lines.
func SearchFiles(raw string) map[string]interface{} {
	var req SearchFilesRequest
	_ = json.Unmarshal([]byte(raw), &req)
	searchTerm := req.Term

	currDir, err := filepath.Abs(".")
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	files, err := os.ReadDir(currDir)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	var output strings.Builder
	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".") || f.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(currDir, f.Name()))
		if err != nil {
			return map[string]interface{}{
				"error": err.Error(),
			}
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.Contains(line, searchTerm) {
				output.WriteString(fmt.Sprintf(`%s:%d:%s\n`, f.Name(), i, line))
			}
		}
	}

	return map[string]interface{}{
		"matched_lines": output.String(),
	}
}

// searchFilesMetadata is the metadata for the search_files function.
var searchFilesMetadata = &openai.FunctionDefine{
	Name:        "search_files",
	Description: "Search for a term in all the files in the current directory",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"term": {
				Type:        openai.JSONSchemaTypeString,
				Description: `string to search for inside files`,
			},
		},
		Required: []string{"term"},
	},
}

// Compile compiles the function in the current directory using the "make build" command.
// It returns a map with the compilation output.
func Compile() map[string]interface{} {
	cmd := exec.Command("make", "build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"error": string(output),
		}
	}

	return map[string]interface{}{
		"output": string(output),
	}
}

// compileMetadata is the metadata for the compile function.
var compileMetadata = &openai.FunctionDefine{
	Name:        "compile",
	Description: "Compile the function in the current directory",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"dummy": {
				Type: openai.JSONSchemaTypeNull,
			},
		},
		Required: []string{},
	},
}
