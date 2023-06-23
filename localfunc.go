package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-skynet/go-llama.cpp"
	"github.com/sashabaranov/go-openai"
)

// EditFilesRequest represents the request structure for editing files.
type EditFilesRequest struct {
	Instructions string `json:"instructions,omitempty"`
	Path         string `json:"path,omitempty"`
}

// EditFiles takes a raw string and a llama model and performs file editing.
func EditFiles(raw string, model *llama.LLama) map[string]string {
	fmt.Println(raw)
	var req EditFilesRequest
	_ = json.Unmarshal([]byte(raw), &req)
	filePath := req.Path

	// Get the current directory path
	currDir, err := filepath.Abs(".")
	if err != nil {
		return map[string]string{
			"error": err.Error(),
		}
	}

	// Clean the file path
	_, err = cleanPaths([]string{filePath}, currDir)
	if err != nil {
		return map[string]string{
			"error": err.Error(),
		}
	}

	// Read the file contents
	contents, err := os.ReadFile(filePath)

	// Create the prompt for the llama model
	prompt := fmt.Sprintf(wizardLMFormat, contents, req.Instructions)

	fmt.Println(prompt)

	var fullResp string

	// Loop until the full response is generated
	for len(fullResp) < len(contents) {
		var resp string
		_, err = model.Predict(prompt,
			llama.Debug,
			llama.SetTokenCallback(func(s string) bool {
				fmt.Print(s)
				resp += s
				return true
			}),
			llama.SetTokens(len(contents)),
			llama.SetThreads(runtime.NumCPU()),
			llama.SetTopK(90),
			llama.SetTopP(0.86),
			llama.SetStopWords("llama"),
		)
		if err != nil {
			return map[string]string{
				"error": err.Error(),
			}
		}

		resp = cleanRes(resp)

		fullResp += resp
		prompt += resp
	}

	// Write the new file contents
	err = os.WriteFile(filePath, []byte(fullResp), 0755)
	if err != nil {
		return map[string]string{
			"error": err.Error(),
		}
	}

	return map[string]string{
		"success": "true",
	}
}

// cleanRes cleans the response by removing unnecessary text.
func cleanRes(in string) string {
	if strings.Contains(in, "```") {
		_, in, _ = strings.Cut(in, "```")
		in, _, _ = strings.Cut(in, "```")
		lines := strings.Split(in, "\n")
		lines = lines[1:]
		in = strings.Join(lines, "\n")
	}

	return in
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
				Type:        openai.JSONSchemaTypeString,
				Description: "JSON-escaped natural language instructions for what modifications to perform on the files.",
			},
		},
		Required: []string{"paths", "instructions"},
	},
}

const alpacaFormat = `### Instruction:

%s

Output the new file contents from the file above, modified using
only the instructions below. Do not make any modifications other
than what was specified in the instructions, and do not output
any text other than the new file contents:

%s

### Response
`

const wizardLMFormat = `
USER: 

%s

Output the new file contents from the file above, modified using
only the instructions below. Do not make any modifications other
than what was specified in the instructions, and do not output
any text other than the new file contents:

%s

ASSISTANT:
`
