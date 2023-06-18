package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
)



type WriteFileRequest struct {
	Path    string `json:"path,omitempty"`
	Content string `json:"content,omitempty"`
}

func WriteFile(raw string) map[string]any {
	var req WriteFileRequest
	_ = json.Unmarshal([]byte(raw), &req)
	if req.Path == "" || req.Content == "" {
		log.Printf("path: %s\n", req.Path)
		log.Printf("content: %s\n", req.Content)
		return map[string]any{
			"error": "must provide path and content",
		}
	}

	currDir, err := filepath.Abs(".")
	e(err)

	fullPaths, err := cleanPaths([]string{req.Path}, currDir)
	if err != nil {
		return map[string]any{
			"error": err.Error(),
		}
	}

	fullPath := fullPaths[0]

	err = os.WriteFile(fullPath, []byte(req.Content), 0755)
	e(err)

	return map[string]any{
		"success": true,
	}
}

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