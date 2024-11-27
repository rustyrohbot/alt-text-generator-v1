package handlers

import (
	"log"
	"net/http"
	"os"

	"alt-text-generator/internal/config"
)

func SaveApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey := r.FormValue("apiKey")
	mode := r.FormValue("mode")

	if apiKey == "" || mode == "" {
		http.Error(w, "API key and mode are required", http.StatusBadRequest)
		return
	}

	var envKey string
	switch mode {
	case "openai":
		envKey = "OPEN_AI_API_KEY"
	case "anthropic":
		envKey = "ANTHROPIC_API_KEY"
	default:
		http.Error(w, "Invalid mode", http.StatusBadRequest)
		return
	}

	// Create or update .env file
	if err := config.UpdateEnvFile(".env", envKey, apiKey); err != nil {
		log.Printf("Error saving API key: %v", err)
		http.Error(w, "Failed to save API key", http.StatusInternalServerError)
		return
	}

	// Set for immediate use
	os.Setenv(envKey, apiKey)

	// Redirect back to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
