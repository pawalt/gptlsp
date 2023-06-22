package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// ListFilesRequest is a struct that represents the request to list files.
type ListFilesRequest struct {
	Directory  string `json:"directory,omitempty"` // Relative path to list files from (default: ".")
	MaxResults int    `json:"max_results,omitempty"` // Maximum number of subtree results to show (default: "20")
}

// ListFiles is a function that lists files based on the provided request.
// It takes a raw JSON string as input and returns a map containing the filepaths.
func ListFiles(raw string) map[string][]string {
	req := ListFilesRequest{
		Directory:  ".",
		MaxResults: 20,
	}
	_ = json.Unmarshal([]byte(raw), &req)
	dir := req.Directory

	currDir, err := filepath.Abs(".")
	if err != nil {
		return map[string][]string{
			"error": []string{err.Error()},
		}
	}

	fullPaths, err := cleanPaths([]string{dir}, currDir)
	if err != nil {
		return map[string][]string{
			"error": []string{err.Error()},
		}
	}

	fullPath := fullPaths[0]

	fileNames, directories, err := exploreDirectory(fullPath, currDir)
	if err != nil {
		return map[string][]string{
			"error": []string{err.Error()},
		}
	}

	// Recursively get directories until we run out of result size.
	for len(directories) > 0 {
		var newFileNames []string
		var newDirectories []string
		for _, dirToExplore := range directories {
			exploredFiles, exploredDirectories, err := exploreDirectory(dirToExplore, currDir)
			if err != nil {
				return map[string][]string{
					"error": []string{err.Error()},
				}
			}

			newFileNames = append(newFileNames, exploredFiles...)
			newDirectories = append(newDirectories, exploredDirectories...)
		}

		if len(newFileNames)+len(fileNames) > req.MaxResults {
			break
		}

		fileNames = append(fileNames, newFileNames...)
		directories = newDirectories
	}

	sort.Strings(fileNames)

	return map[string][]string{
		"filepaths": fileNames,
	}
}

// exploreDirectory is a helper function that explores a directory and returns the file names and directories.
func exploreDirectory(dir string, root string) ([]string, []string, error) {
	relativeBase := strings.TrimPrefix(dir, root)
	relativeBase = strings.TrimPrefix(relativeBase, "/")

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}

	var fileNames []string
	var directories []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}

		fName := filepath.Join(relativeBase, f.Name())
		if f.IsDir() {
			fName += "/"
			directories = append(directories, filepath.Join(dir, fName))
		}

		fileNames = append(fileNames, fName)
	}

	return fileNames, directories, nil
}

var listFilesMetadata = &openai.FunctionDefine{
	Name:        "list_files",
	Description: "List files in the specified directory",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"directory": {
				Type:        openai.JSONSchemaTypeString,
				Description: `Relative path to list files from (default: ".")`,
			},
			"max_results": {
				Type:        openai.JSONSchemaTypeNumber,
				Description: `Maximum number of subtree results to show (default: "20")`,
			},
		},
		Required: []string{},
	},
}
