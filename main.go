package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

// main is the entry point of the program
func main() {
	runRepl()
}

// runRepl runs the REPL (Read-Eval-Print Loop)
func runRepl() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")

	// Create a new client with the API key
	client := openai.NewClient(apiKey)

	// Start the REPL with the client
	repl(client)
}