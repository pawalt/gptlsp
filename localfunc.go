package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-skynet/go-llama.cpp"
	"github.com/sashabaranov/go-openai"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

type EditFilesRequest struct {
	Instructions string `json:"instructions,omitempty"`
	Path string `json:"path,omitempty"`
}

func EditFiles(raw string, model *llama.LLama) map[string]string {
	fmt.Println(raw)
	var req EditFilesRequest
	_ = json.Unmarshal([]byte(raw), &req)
	filePath := req.Path

	currDir, err := filepath.Abs(".")
	e(err)

	_, err = cleanPaths([]string{filePath}, currDir)
	if err != nil {
		return map[string]string{
			"error": err.Error(),
		}
	}

	contents, err := os.ReadFile(filePath)

	words := strings.FieldsFunc(string(contents), func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})

	prompt := fmt.Sprintf(`### Instruction:

%s

Read the inputted file contents and output the new desired file contents
according to the instructions:

%s

### Response
`, contents, req.Instructions)

	var resp string
	_, err = model.Predict(prompt,
		 llama.Debug,
		 llama.SetTokenCallback(func(s string) bool {
			fmt.Print(s)
			resp += s
			return true
		}),
		llama.SetTokens(len(words) + 256),
		llama.SetThreads(runtime.NumCPU()),
		llama.SetTopK(90),
		llama.SetTopP(0.86),
		llama.SetStopWords("llama"),
	)
	e(err)

	err = os.WriteFile(filePath, []byte(resp), 0755)
	e(err)

	return map[string]string{
		"success": "true",
	}
}

var editFilesMetadata = &openai.FunctionDefine{
	Name:        "edit_files",
	Description: "Send instructions for editing a set of files.",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"path": {
				Type:        openai.JSONSchemaTypeString,
				Description: `Relative path to the file to read and write`,
			},
			"instructions": {
				Type: openai.JSONSchemaTypeString,
				Description: "JSON-escaped natural language instructions for what modifications to perform on the files.",
			},
		},
		Required: []string{"paths", "instructions"},
	},
}
