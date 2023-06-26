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
	orca7      = "./models/orca-mini-7b.ggmlv3.q4_K_M.bin"
	orca13     = "./models/orca-mini-13b.ggmlv3.q4_K_M.bin"
	orca135    = "./models/orca-mini-13b.ggmlv3.q5_K_M.bin"
)

const wizardBaseFormat = ` You are a helpful assistant.

You have sent the following messages already:
%s

Now respond to the user.

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
	l, err := llama.New(orca135, llama.EnableF16Memory, llama.SetContext(2048), llama.SetGPULayers(1))
	if err != nil {
		fmt.Println("Loading the model failed:", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Model loaded successfully.\n")

	runRefactorTest(l)
}

func runRefactorTest(l *llama.LLama) {
	prompt := fmt.Sprintf(`Sure, here's a new prompt that would guide the AI to refactor the code accurately and only return the updated code:

### System:
You are an AI assistant that understands programming, code refactoring, and is extremely adept at following instructions. Help refactor the code from using the standard "error" type to "cerr.Cerr" type. Your main task is to only return the refactored code.

### User:

Refactor the function error types to cerr.Cerr. Update GRPC handlers to return cerr.Cerr instead of error, and only return the updated text.

Modify as such:

Before:
func (s *Server) X(ctx context.Context, req *Y) (*Z, error)
After:
func (s *Server) X(ctx context.Context, req *Y) (*Z, cerr.Cerr)

Errors should return using cerr wrappers.

For instance:

Validation errors:

Before:
if len(users) == 0 {
return nil, errors.New("cannot have no users")
}
After:
if len(users) == 0 {
return nil, cerr.New(ctx, "cannot have no users")
}

Library errors:

Before:
_, err := someOp()
if err != nil {
	return nil, err
}
After:
_, err := someOp()
if err != nil {
	return nil, cerr.Wrap(ctx, "op failed", err)
}

### Input:

func (s *Server) AddMethods(ctx context.Context, req *AddRequest) (*AddResponse, error) {
	res, err := add(req.X, req.Y)
	if err != nil {
		return nil, err
	}
	return &AddResponse{res}, nil
}

func add (x, y int) (int, error) {
	return x + y, nil
}

### Response:
`)
	_, err := l.Predict(prompt, llama.Debug, llama.SetTokenCallback(func(token string) bool {
		fmt.Print(token)
		return true
	}), llama.SetTokens(512), llama.SetThreads(threads), llama.SetTopK(90), llama.SetTopP(0.86), llama.SetStopWords("llama"))
	if err != nil {
		log.Panic(err)
	}
}

func runLocalRepl(l *llama.LLama) {
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

		text := fmt.Sprintf(wizardBaseFormat, textMessages, input)

		history = append(history, msg{
			role:    "human",
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
