package main

import (
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

var refactoringFunctions = []*openai.FunctionDefine{
	searchFilesMetadata,
	refactorFilesMetadata,
	compileMetadata,
	switchModeMetadata,
}

var codingFunctions = []*openai.FunctionDefine{
	searchFilesMetadata,
	modifyFilesMetadata,
	compileMetadata,
	switchModeMetadata,
}

type SwitchModeRequest struct {
	Mode string `json:"mode,omitempty"`
}

func SwitchMode(raw string) map[string]any {
	var req SwitchModeRequest
	_ = json.Unmarshal([]byte(raw), &req)
	if req.Mode == "" {
		return map[string]any{
			"error": "must provide a mode",
		}
	}

	switch req.Mode {
	case "refactor":
		activeFunctions = refactoringFunctions
	case "code":
		activeFunctions = codingFunctions
	default:
		return map[string]any{
			"error": fmt.Sprintf("unrecognized mode: %s", req.Mode),
		}
	}

	return map[string]any{
		"success": true,
	}
}

var switchModeMetadata = &openai.FunctionDefine{
	Name:        "switch_mode",
	Description: "Switch the agent's mode of operation",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"mode": {
				Type:        openai.JSONSchemaTypeString,
				Enum:        []string{"refactor", "code"},
				Description: `Mode to switch to`,
			},
		},
		Required: []string{"mode"},
	},
}
