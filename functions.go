package main

import (
	"time"

	"github.com/sashabaranov/go-openai"
)

func GetTime() string {
	return time.Now().String()
}

var getTimeMetadata = &openai.FunctionDefine{
	Name:        "get_time",
	Description: "Get the current time",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"dummy": {
				Type:        openai.JSONSchemaTypeNull,
				Description: "",
			},
		},
		Required: []string{},
	},
}
