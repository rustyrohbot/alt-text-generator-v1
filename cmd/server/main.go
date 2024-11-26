package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"alt-text-generator/internal/api"
	"alt-text-generator/internal/config"
	"alt-text-generator/internal/handlers"
)

func main() {
	// Define flags for selecting which API to use
	useOpenAI := flag.Bool("openai", false, "Use OpenAI API")
	useAnthropic := flag.Bool("anthropic", false, "Use Anthropic API")
	flag.Parse()

	// Load environment variables from .env file
	log.Println("Loading environment variables from .env file")
	err := config.LoadEnvFile(".env")
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file: %v", err)
	}
	log.Println("Successfully loaded .env file")

	// Set the appropriate API selection function and mode
	var generateAltTextFunc func(string) (string, error)
	var mode string
	if *useOpenAI {
		generateAltTextFunc = api.GenerateAltTextOpenAI
		mode = "openai"
	} else if *useAnthropic {
		generateAltTextFunc = api.GenerateAltTextClaude
		mode = "anthropic"
	} else {
		log.Fatalf("You must specify either -openai or -anthropic flag.")
	}

	// Set up routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HomeHandler(w, r, mode)
	})
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		handlers.UploadHandler(w, r, generateAltTextFunc, mode)
	})
	http.HandleFunc("/saveApiKey", handlers.SaveApiKeyHandler)

	// Start server
	port := ":8080"
	fmt.Printf("Starting server on %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
