package main

import (
	"time"

	"github.com/franciscoescher/goopenai"
)

func GetTime() string {
	return time.Now().String()
}

var getTimeMetadata = goopenai.Function{
	Name:        "get_time",
	Description: "Get the current time",
	Parameters: goopenai.U{
		"type":       "object",
		"properties": goopenai.U{},
	},
}
