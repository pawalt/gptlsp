package main

import (
	"time"

	"github.com/sashabaranov/go-openai"
)

// GetTime retrieves the current time and returns it in a map
func GetTime() map[string]string {
	return map[string]string{
		"time": time.Now().String(),
	}
}

var getTimeMetadata = &openai.FunctionDefine{
	Name:        "get_time",
	Description: "Get the current time",
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