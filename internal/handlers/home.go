package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"alt-text-generator/internal/types"
)

var tmpl = template.Must(template.ParseFiles(filepath.Join("web", "template.html")))

func HomeHandler(w http.ResponseWriter, r *http.Request, mode string) {
	log.Println("Serving home page")

	// Check if API key exists
	var apiKey string
	if mode == "openai" {
		apiKey = os.Getenv("OPEN_AI_API_KEY")
	} else {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	data := types.TemplateData{
		Mode:          mode,
		APIKeyMissing: apiKey == "",
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		renderHomeError(w, err.Error())
		return
	}
}

func renderHomeError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `
		<div class="bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded-lg">
			<p class="font-bold">Error loading page: %s</p>
			<button onclick="window.location.reload()" class="mt-2 bg-red-100 text-red-700 px-4 py-2 rounded hover:bg-red-200">
				Reload Page
			</button>
		</div>
	`, message)
}
