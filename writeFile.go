package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
)

// WriteFileRequest is a struct that specifies the parameters for writing to a file.
type WriteFileRequest struct {
	Path    string `json:"path,omitempty"`    // Path is the relative path to the file to write to.
	Content string `json:"content,omitempty"` // Content is the JSON-escaped content to write to the file.
}

// WriteFile writes content to a file at the given path.
func WriteFile(raw string) map[string]interface{} {
	var req WriteFileRequest
	_ = json.Unmarshal([]byte(raw), &req)
	if req.Path == "" || req.Content == "" {
		log.Printf("path: %s\n", req.Path)
		log.Printf("content: %s\n", req.Content)
		return map[string]interface{}{
			"error": "must provide path and content",
		}
	}

	currDir, err := filepath.Abs(".")
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	fullPaths, err := cleanPaths([]string{req.Path}, currDir)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	fullPath := fullPaths[0]

	err = os.WriteFile(fullPath, []byte(req.Content), 0755)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
	}
}

// writeFileMetadata is a variable that holds the metadata for the WriteFile function.
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
		Required: []string{"path"},
	},
}