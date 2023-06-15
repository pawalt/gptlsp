package main

import (
	"log"
	"os"

	"github.com/franciscoescher/goopenai"
	"github.com/joho/godotenv"
)

const (
	TURBO_WITH_FUNCTIONS = "gpt-3.5-turbo-0613"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	organization := os.Getenv("API_ORG")

	client := goopenai.NewClient(apiKey, organization)

	repl(client)
}
