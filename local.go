package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/go-skynet/go-llama.cpp"
)

var (
	threads = runtime.NumCPU()
)

const (
	wizard134  = "./models/wizardLM-13B-Uncensored.ggmlv3.q4_1.bin"
	wizard74   = "./models/wizardLM-7B.ggmlv3.q4_1.bin"
	nousHermes = "./models/nous-hermes-13b.ggmlv3.q4_1.bin"
)

const wizardBaseFormat = `
USER: %s
ASSISTANT:
`

const alpacaBaseFormat = `### Instruction:

You are chatbot who has sent the following messages:
%s

The user has asked you the following.

%s
### Response:
AI: `

type msg struct {
	role    string
	content string
}

// runLocal is the main function that runs the program locally.
func runLocal() {
	// Load the model with the specified options.
	l, err := llama.New(nousHermes, llama.EnableF16Memory, llama.SetContext(2048), llama.SetGPULayers(1))
	if err != nil {
		fmt.Println("Loading the model failed:", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Model loaded successfully.\n")

	reader := bufio.NewReader(os.Stdin)

	var history []msg
	for {
		textMessages := ""
		for _, m := range history {
			textMessages += fmt.Sprintf("%s: %s\n", strings.ToUpper(m.role), m.content)
		}

		input, err := readInput(reader)
		if err != nil {
			log.Panic(err)
		}

		text := fmt.Sprintf(alpacaBaseFormat, textMessages, input)

		history = append(history, msg{
			role:    "user",
			content: strings.TrimSpace(input),
		})

		resp, err := l.Predict(text, llama.Debug, llama.SetTokenCallback(func(token string) bool {
			fmt.Print(token)
			return true
		}), llama.SetTokens(256), llama.SetThreads(threads), llama.SetTopK(90), llama.SetTopP(0.86), llama.SetStopWords("llama"))
		if err != nil {
			log.Panic(err)
		}
		history = append(history, msg{
			role:    "AI: ",
			content: resp,
		})

		fmt.Printf("\n\n")
	}
}

// readInput reads input until an empty line is entered.
func readInput(reader *bufio.Reader) (string, error) {
	fmt.Print("user: ")
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return line, nil
}
