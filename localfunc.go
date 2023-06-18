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

	prompt := fmt.Sprintf(wizardLMFormat, contents, req.Instructions)

	fmt.Println(prompt)

	var fullResp string

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
		e(err)

		resp = cleanRes(resp)

		fullResp += resp
		prompt += resp
	}

	err = os.WriteFile(filePath, []byte(fullResp), 0755)
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

const alpacaFormat = `### Instruction:

%s

Output the new file contents from the file above, modified using
only the instructions below. Do not make any modifications other
than what was specified in the instructions, and do not output
any text other than the new file contents:

%s

### Response
`

const wizardLMFormat = `You are a helpful assistant.
USER: 

%s

Output the new file contents from the file above, modified using
only the instructions below. Do not make any modifications other
than what was specified in the instructions, and do not output
any text other than the new file contents:

%s

ASSISTANT:`
