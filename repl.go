package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-skynet/go-llama.cpp"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

func repl(client *openai.Client) {
	var messages []openai.ChatCompletionMessage

	hermesModel := "./models/nous-hermes-13b.ggmlv3.q4_1.bin"
	//wizardLM := "./models/WizardLM-7B-uncensored.ggmlv3.q4_0.bin"
	l, err := llama.New(hermesModel, llama.EnableF16Memory, llama.SetContext(2048), llama.EnableEmbeddings, llama.SetGPULayers(1))
	if err != nil {
		log.Panicf("Loading the model failed: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	inputRequired := "user"
	functionName := ""
	functionArgs := ""
	for {
		var nextMessage openai.ChatCompletionMessage
		if inputRequired == "user" {
			fmt.Print("user: ")
			scanner.Scan()
			nextMessage = openai.ChatCompletionMessage{
				Role:    "user",
				Content: scanner.Text(),
			}
		} else if inputRequired == "function_call" {
			nextMessage = executeFunction(client, functionName, functionArgs, messages, l)
		} else {
			log.Panicf("bad input required %s", inputRequired)
		}

		messages = append(messages, nextMessage)

		// model := openai.GPT3Dot5Turbo0613
		model := openai.GPT40613
		completions, err := createChatCompletion(client, model, messages, []*openai.FunctionDefine{
			globFilesMetadata,
			getGoSymbolsMetadata,
			editFilesMetadata,
		})
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
