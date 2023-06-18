package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func main() {
	runRepl()
}

func runRepl() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(apiKey)

	repl(client)
}

func e(err error) {
	if err != nil {
		log.Panic(err)
	}
}