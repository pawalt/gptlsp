package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

func repl(client *openai.Client) {
	var messages []openai.ChatCompletionMessage

	scanner := bufio.NewScanner(os.Stdin)
	inputRequired := "user"
	functionName := ""
	//functionArgs := ""
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
			switch functionName {
			case getTimeMetadata.Name:
				nextMessage = openai.ChatCompletionMessage{
					Role:    "function",
					Name:    functionName,
					Content: GetTime(),
				}
			default:
				log.Panicf("unrecognized function name %s", functionName)
			}
			fmt.Printf("%s: %s\n", nextMessage.Name, nextMessage.Content)
		} else {
			log.Panicf("bad input required %s", inputRequired)
		}

		if nextMessage.Role == "" {
			log.Panic("why no role???")
		}

		messages = append(messages, nextMessage)

		r := openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo0613,
			Messages:    messages,
			Temperature: 0.7,
			Functions: []*openai.FunctionDefine{
				getTimeMetadata,
			},
		}

		completions, err := client.CreateChatCompletion(context.Background(), r)
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
			//functionArgs = choice.Message.FunctionCall.Arguments
		default:
			fmt.Printf("assistant: %s\n", choice.Message.Content)
			inputRequired = "user"
		}
	}
}

func s(str string) *string {
	return &str
}
