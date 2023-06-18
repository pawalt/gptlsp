package main

import (
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"path/filepath"
	"strings"
)


type GlobFilesRequest struct {
	Pattern string `json:"pattern,omitempty"`
}

func GlobFiles(raw string) map[string][]string {
	var req GlobFilesRequest
	_ = json.Unmarshal([]byte(raw), &req)
	pattern := req.Pattern

	// Get the absolute path of the current directory
	currDir, err := filepath.Abs(".")
	if err != nil {
		return map[string][]string{
			"error": {"could not determine current directory"},
		}
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return map[string][]string{}
	}

	var validMatches []string
	for _, match := range matches {
		absMatch, err := filepath.Abs(match)
		if err != nil {
			return map[string][]string{
				"error": {fmt.Sprintf("could not determine absolute path of match: %v", err)},
			}
		}

		// Check if the match is in the current directory or a subdirectory
		if strings.HasPrefix(absMatch, currDir) {
			validMatches = append(validMatches, match)
		}
	}

	return map[string][]string{
		"filepaths": validMatches,
	}
}

var globFilesMetadata = &openai.FunctionDefine{
	Name:        "glob_files",
	Description: "Get a list of files matching the glob pattern",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"pattern": {
				Type:        openai.JSONSchemaTypeString,
				Description: `Glob pattern to match files`,
			},
		},
		Required: []string{"pattern"},
	},
}