package handlers

import (
	"log"
	"net/http"
	"os"

	"alt-text-generator/internal/components"
)

func HomeHandler(w http.ResponseWriter, r *http.Request, mode string) {
	log.Println("Serving home page")

	// Check if API key exists
	var apiKey string
	if mode == "openai" {
		apiKey = os.Getenv("OPEN_AI_API_KEY")
	} else {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	component := components.Home(mode, apiKey == "")
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}
