package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/franciscoescher/goopenai"
)

func repl(client *goopenai.Client) {
	messages := []goopenai.Message{}

	scanner := bufio.NewScanner(os.Stdin)
	inputRequired := "user"
	functionName := ""
	//functionArgs := ""
	for {
		var nextMessage goopenai.Message
		if inputRequired == "user" {
			fmt.Print("user: ")
			scanner.Scan()
			nextMessage = goopenai.Message{
				Role:    "user",
				Content: s(scanner.Text()),
			}
		} else if inputRequired == "function_call" {
			switch functionName {
			case getTimeMetadata.Name:
				nextMessage = goopenai.Message{
					Role:    "function",
					Name:    functionName,
					Content: s(GetTime()),
				}
			default:
				log.Panicf("unrecognized function name %s", functionName)
			}
			fmt.Printf("%s: %s\n", nextMessage.Name, *nextMessage.Content)
		} else {
			log.Panicf("bad input required %s", inputRequired)
		}

		if nextMessage.Role == "" {
			log.Panic("why no role???")
		}

		messages = append(messages, nextMessage)

		r := goopenai.CreateCompletionsRequest{
			Model:       TURBO_WITH_FUNCTIONS,
			Messages:    messages,
			Temperature: 0.7,
			Functions: []goopenai.Function{
				getTimeMetadata,
			},
		}

		completions, err := client.CreateCompletions(context.Background(), r)
		if err != nil {
			log.Panicf("could not create completions %v", err)
		}
		if completions.Error.Message != "" {
			log.Panicf("could not create completions %v", completions.Error.Message)
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
			fmt.Printf("assistant: %s\n", *choice.Message.Content)
			inputRequired = "user"
		}
	}
}

func s(str string) *string {
	return &str
}
