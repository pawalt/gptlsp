// Package main provides functions to list files in a directory.

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// ListFilesRequest represents the request parameters for listing files.
type ListFilesRequest struct {
	Directory  string `json:"directory,omitempty"`  // Relative path to the directory (default: ".")
	MaxResults int    `json:"max_results,omitempty"` // Maximum number of results to show (default: 20)
}

// ListFiles lists the files in the specified directory.
// It takes a raw JSON string as input and returns a map with the file paths.
func ListFiles(raw string) map[string][]string {
	// Parse the raw JSON string into a ListFilesRequest struct
	req := ListFilesRequest{
		Directory:  ".",
		MaxResults: 20,
	}
	_ = json.Unmarshal([]byte(raw), &req)

	// Get the absolute path of the current directory
	currDir, err := filepath.Abs(".")
	if err != nil {
		return map[string][]string{
			"error": []string{err.Error()},
		}
	}

	// Clean the input directory path and get the full path
	fullPaths, err := cleanPaths([]string{req.Directory}, currDir)
	if err != nil {
		return map[string][]string{
			"error": []string{err.Error()},
		}
	}
	fullPath := fullPaths[0]

	// Explore the directory and get the file names and subdirectories
	fileNames, directories, err := exploreDirectory(fullPath, currDir)
	if err != nil {
		return map[string][]string{
			"error": []string{err.Error()},
		}
	}

	// Explore the subdirectories recursively until the maximum number of results is reached
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

		// Break the loop if the total number of files exceeds the maximum results
		if len(newFileNames)+len(fileNames) > req.MaxResults {
			break
		}

		fileNames = append(fileNames, newFileNames...)
		directories = newDirectories
	}

	// Sort the file names alphabetically
	sort.Strings(fileNames)

	return map[string][]string{
		"filepaths": fileNames,
	}
}

// exploreDirectory explores a directory and returns the file names and subdirectories.
// It takes the directory path and the root path as input, and returns the file names, subdirectories, and an error if any.
func exploreDirectory(dir string, root string) ([]string, []string, error) {
	// Get the relative path of the directory
	relativeBase := strings.TrimPrefix(dir, root)
	relativeBase = strings.TrimPrefix(relativeBase, "/")

	// Read the directory and get the file names and subdirectories
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}

	var fileNames []string
	var directories []string
	for _, f := range files {
		// Skip hidden files and directories starting with "."
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

// listFilesMetadata defines the metadata for the ListFiles function.
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