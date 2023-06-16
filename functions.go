package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
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
				Type: openai.JSONSchemaTypeNull,
			},
		},
		Required: []string{},
	},
}

type ListFilesRequest struct {
	Directory  string `json:"directory,omitempty"`
	MaxResults int    `json:"max_results,omitempty"`
}

func ListFiles(raw string) any {
	req := ListFilesRequest{
		Directory:  ".",
		MaxResults: 20,
	}
	_ = json.Unmarshal([]byte(raw), &req)
	dir := req.Directory

	currDir, err := filepath.Abs(".")
	e(err)

	fullPaths, err := cleanPaths([]string{dir}, currDir)
	if err != nil {
		return map[string]string{
			"error": err.Error(),
		}
	}

	fullPath := fullPaths[0]

	fileNames, directories, err := exploreDirectory(fullPath, currDir)
	e(err)

	// Recursively get directories until we run out of result size.
	for len(directories) > 0 {
		var newFileNames []string
		var newDirectories []string
		for _, dirToExplore := range directories {
			exploredFiles, exploredDirectories, err := exploreDirectory(dirToExplore, currDir)
			e(err)

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

func e(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
