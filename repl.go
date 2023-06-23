// Package main is the main package of the program.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-skynet/go-llama.cpp"

	"github.com/sashabaranov/go-openai"
)

const (
	gpt3 = openai.GPT3Dot5Turbo0613
	gpt4 = openai.GPT40613
)

var activeMode = "refactor"

// repl is the main read-eval-print loop function.
func repl(client *openai.Client) {
	var messages []openai.ChatCompletionMessage

	scanner := bufio.NewScanner(os.Stdin)
	inputRequired := "user"
	functionName := ""
	functionArgs := ""
	for {
		var nextMessage openai.ChatCompletionMessage
		if inputRequired == "user" {
			fmt.Printf("(%s) user: ", activeMode)
			scanner.Scan()
			nextMessage = openai.ChatCompletionMessage{
				Role:    "user",
				Content: scanner.Text(),
			}
		} else if inputRequired == "function_call" {
			nextMessage = executeFunction(client, functionName, functionArgs, messages, nil)
		} else {
			log.Panicf("bad input required %s", inputRequired)
		}

		messages = append(messages, nextMessage)

		completions, err := createChatCompletion(client, gpt4, messages, modeToFunctions[activeMode])
		if err != nil {
			log.Panicf("could not create completions %v", err)
		}

		if len(completions.Choices) != 1 {
			log.Panicf("somehow not length 1 choices %v", completions.Choices)
		}

		choice := completions.Choices[0]

		messages = append(messages, choice.Message)

		switch choice.FinishReason {
		case "function_call":
			inputRequired = "function_call"
			functionName = choice.Message.FunctionCall.Name
			fmt.Printf("assistant: calling function %s\n", functionName)
			functionArgs = choice.Message.FunctionCall.Arguments
		default:
			fmt.Printf("assistant: %s\n", choice.Message.Content)
			inputRequired = "user"
		}
	}
}

// executeFunction executes the specified function with the given arguments.
func executeFunction(client *openai.Client, functionName string, functionArgs string, history []openai.ChatCompletionMessage, llamaModel *llama.LLama) openai.ChatCompletionMessage {
	var resp any
	switch functionName {
	case getTimeMetadata.Name:
		resp = GetTime()
	case listFilesMetadata.Name:
		resp = ListFiles(functionArgs)
	case readFilesMetadata.Name:
		resp = ReadFiles(functionArgs)
	case writeFileMetadata.Name:
		resp = WriteFile(functionArgs)
	case analyzeMetadata.Name:
		resp = Analyze(functionArgs, client, history)
	case modifyFilesMetadata.Name:
		resp = Modify(functionArgs, client, history)
	case globFilesMetadata.Name:
		resp = GlobFiles(functionArgs)
	case getGoSymbolsMetadata.Name:
		resp = GetGoSymbols(functionArgs)
	case editFilesMetadata.Name:
		resp = EditFiles(functionArgs, llamaModel)
	case refactorFilesMetadata.Name:
		resp = Refactor(functionArgs, client)
	case searchFilesMetadata.Name:
		resp = SearchFiles(functionArgs)
	case compileMetadata.Name:
		resp = Compile()
	case switchModeMetadata.Name:
		resp = SwitchMode(functionArgs)
	default:
		log.Panicf("unrecognized function name %s", functionName)
	}
	b, err := json.Marshal(resp)
	if err != nil {
		log.Panicf("err marshalling %v", err)
	}

	respText := string(b)
	humanReadable := respText

	truncateLen := 200
	if len(humanReadable) > truncateLen {
		humanReadable = humanReadable[:truncateLen] + "..."
	}

	fmt.Printf("%s(%s): %s\n", functionName, functionArgs, humanReadable)

	return openai.ChatCompletionMessage{
		Role:    "function",
		Name:    functionName,
		Content: respText,
	}
}
