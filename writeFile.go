package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
)

// WriteFileRequest represents the request structure for the WriteFile function
type WriteFileRequest struct {
	Path    string `json:"path,omitempty"`    // Path represents the relative path to the file to write to
	Content string `json:"content,omitempty"` // Content represents the JSON-escaped content to write to the file
}

// WriteFile writes content to a file at the specified path
// Input: raw (string) - the raw JSON request string
// Output: map[string]interface{} - a map representing the response
func WriteFile(raw string) map[string]interface{} {
	var req WriteFileRequest

	// Unmarshal the raw JSON request string into the WriteFileRequest struct
	_ = json.Unmarshal([]byte(raw), &req)

	// Check if the path or content is empty
	if req.Path == "" || req.Content == "" {
		log.Printf("path: %s\n", req.Path)
		log.Printf("content: %s\n", req.Content)
		return map[string]interface{}{
			"error": "must provide path and content",
		}
	}

	// Get the current directory
	currDir, err := filepath.Abs(".")
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Clean the provided path and get the full path
	fullPaths, err := cleanPaths([]string{req.Path}, currDir)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	fullPath := fullPaths[0]

	// Write the content to the file
	err = os.WriteFile(fullPath, []byte(req.Content), 0755)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Return success response
	return map[string]interface{}{
		"success": true,
	}
}

// Define metadata for the WriteFile function
var writeFileMetadata = &openai.FunctionDefine{
	Name:        "write_file",
	Description: "Write to file at the given path",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"path": {
				Type:        openai.JSONSchemaTypeString,
				Description: `Relative path to the file to write to`,
			},
			"content": {
				Type:        openai.JSONSchemaTypeString,
				Description: `JSON-escaped content to write to the file`,
			},
		},
		Required: []string{"path", "content"},
	},
}