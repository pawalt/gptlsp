package main

import (
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

var modeToFunctions = map[string][]*openai.FunctionDefine{
	"refactor": {
		searchFilesMetadata,
		refactorFilesMetadata,
		globFilesMetadata,
		switchModeMetadata,
	},
	"code": {
		searchFilesMetadata,
		modifyFilesMetadata,
		globFilesMetadata,
		compileMetadata,
		switchModeMetadata,
	},
}

// Define the SwitchModeRequest struct
type SwitchModeRequest struct {
	Mode string `json:"mode,omitempty"`
}

// SwitchMode function switches the mode of operation for the agent
func SwitchMode(raw string) map[string]interface{} {
	var req SwitchModeRequest
	_ = json.Unmarshal([]byte(raw), &req)
	if req.Mode == "" {
		return map[string]interface{}{
			"error": "must provide a mode",
		}
	}

	if _, found := modeToFunctions[req.Mode]; !found {
		return map[string]interface{}{
			"error": fmt.Sprintf("unrecognized mode: %s", req.Mode),
		}
	}

	activeMode = req.Mode

	return map[string]interface{}{
		"success": true,
	}
}

// Define the switch mode metadata
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
