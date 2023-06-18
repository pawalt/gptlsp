package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)



type ReadFileRequest struct {
	Paths []string `json:"paths,omitempty"`
}

func ReadFiles(raw string) map[string]string {
	var req ReadFileRequest
	_ = json.Unmarshal([]byte(raw), &req)
	filePaths := req.Paths

	currDir, err := filepath.Abs(".")
	e(err)

	cleanedPaths, err := cleanPaths(filePaths, currDir)
	if err != nil {
		return map[string]string{
			"error": err.Error(),
		}
	}

	ret := make(map[string]string)
	for _, fileName := range cleanedPaths {
		contents, err := os.ReadFile(fileName)
		if err != nil {
			return map[string]string{
				"error": fmt.Sprintf("could not read file: %v", err),
			}
		}

		fileName = strings.TrimPrefix(fileName, currDir)
		fileName = strings.TrimPrefix(fileName, "/")

		ret[fileName] = string(contents)
	}

	return ret
}

func cleanPaths(paths []string, root string) ([]string, error) {
	var ret []string
	for _, fp := range paths {
		cleanPath := filepath.Clean(fp)
		fullPath := filepath.Join(root, cleanPath)
		if !strings.HasPrefix(fullPath, root) {
			return nil, fmt.Errorf("invalid path %s escapes base directory", fp)
		}

		ret = append(ret, fullPath)
	}
	return ret, nil
}

var readFilesMetadata = &openai.FunctionDefine{
	Name:        "read_files",
	Description: "Read files at the given paths",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"paths": {
				Type:        openai.JSONSchemaTypeArray,
				Description: `Relative paths to the files to read`,
				Items: &openai.JSONSchemaDefine{
					Type: openai.JSONSchemaTypeString,
				},
			},
		},
		Required: []string{"paths"},
	},
}