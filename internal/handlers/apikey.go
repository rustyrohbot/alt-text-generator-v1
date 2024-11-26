package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"alt-text-generator/internal/config"
)

func SaveApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderApiError(w, "Method not allowed")
		return
	}

	apiKey := r.FormValue("apiKey")
	mode := r.FormValue("mode")

	if apiKey == "" || mode == "" {
		renderApiError(w, "API key and mode are required")
		return
	}

	var envKey string
	switch mode {
	case "openai":
		envKey = "OPEN_AI_API_KEY"
	case "anthropic":
		envKey = "ANTHROPIC_API_KEY"
	default:
		renderApiError(w, "Invalid mode")
		return
	}

	// Create or append to .env file
	f, err := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening .env file: %v", err)
		renderApiError(w, "Failed to save API key")
		return
	}
	defer f.Close()

	// Check if the key already exists
	if err := config.LoadEnvFile(".env"); err == nil {
		existingKey := os.Getenv(envKey)
		if existingKey != "" {
			// Key exists, update it
			if err := config.UpdateEnvFile(".env", envKey, apiKey); err != nil {
				log.Printf("Error updating API key: %v", err)
				renderApiError(w, "Failed to update API key")
				return
			}
		} else {
			// Key doesn't exist, append it
			if _, err := f.WriteString(fmt.Sprintf("%s=%s\n", envKey, apiKey)); err != nil {
				log.Printf("Error writing API key: %v", err)
				renderApiError(w, "Failed to save API key")
				return
			}
		}
	}

	// Set the environment variable for immediate use
	os.Setenv(envKey, apiKey)

	// Redirect back to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func renderApiError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<div class="bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded-lg">
			<p class="font-bold">Error: %s</p>
			<button onclick="window.location.href='/'" class="mt-2 bg-red-100 text-red-700 px-4 py-2 rounded hover:bg-red-200">
				Try Again
			</button>
		</div>
	`, message)
}
