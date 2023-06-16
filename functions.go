package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

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
				Type:        openai.JSONSchemaTypeNull,
			},
		},
		Required: []string{},
	},
}

type ListFilesRequest struct {
	Directory string `json:"directory,omitempty"`
}

func ListFiles(raw string) map[string][]string {
	var resp ListFilesRequest
	_ = json.Unmarshal([]byte(raw), &resp)
	dir := resp.Directory
	if dir == "" {
		dir = "."
	}

	currDir, err := filepath.Abs(".")
	e(err)

	cleanDir := filepath.Clean(dir)
	fullPath := filepath.Join(currDir, cleanDir)
	if !strings.HasPrefix(fullPath, currDir) {
		log.Printf("invalid path %s escapes base directory", dir)
	}

	relativeBase := strings.TrimPrefix(fullPath, currDir)

	files, err := os.ReadDir(fullPath)
	e(err)

	var fileNames []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".")  {
			continue
		}

		fName := relativeBase + f.Name()
		if f.IsDir() {
			fName += "/"
		}

		fileNames = append(fileNames, fName)
	}

	return map[string][]string{
		"filepaths": fileNames,
	}
}

var listFilesMetadata = &openai.FunctionDefine{
	Name:        "list_files",
	Description: "List files in the specified directory",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"directory": {
				Type: openai.JSONSchemaTypeString,
				Description: `Relative path to list files from (default: ".")`,
			},
		},
		Required: []string{},
	},
}


type ReadFileRequest struct {
	Path string `json:"path,omitempty"`
}

func ReadFile(raw string) map[string]string {
	var resp ReadFileRequest
	_ = json.Unmarshal([]byte(raw), &resp)
	fp := resp.Path
	if fp == "" {
		fp = "."
	}

	currDir, err := filepath.Abs(".")
	e(err)

	cleanDir := filepath.Clean(fp)
	fullPath := filepath.Join(currDir, cleanDir)
	if !strings.HasPrefix(fullPath, currDir) {
		return map[string]string{
			"error": fmt.Sprintf("invalid path %s escapes base directory", fp),
		}
	}

	contents, err := os.ReadFile(fullPath)
	if err != nil {
		return map[string]string{
			"error": fmt.Sprintf("could not read file: %v", err),
		}
	}


	return map[string]string{
		"contents": string(contents),
	}
}

var readFileMetadata = &openai.FunctionDefine{
	Name:        "read_file",
	Description: "Read a file at the given path",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"path": {
				Type: openai.JSONSchemaTypeString,
				Description: `Relative path to the file to read`,
			},
		},
		Required: []string{"path"},
	},
}

func e(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
