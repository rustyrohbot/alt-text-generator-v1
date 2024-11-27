package handlers

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"alt-text-generator/internal/components"
)

func UploadHandler(w http.ResponseWriter, r *http.Request, generateAltTextFunc func(string) (string, error), mode string) {
	log.Println("Received upload request")

	if r.Method != http.MethodPost {
		log.Println("Invalid request method. Expected POST.")
		renderError(w, r, "Invalid request method")
		return
	}

	// Verify API key exists before processing upload
	var apiKey string
	if mode == "openai" {
		apiKey = os.Getenv("OPEN_AI_API_KEY")
	} else {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	if apiKey == "" {
		renderError(w, r, "API key not configured")
		return
	}

	// Try to parse the multipart form with a 6MB limit
	if err := r.ParseMultipartForm(6 * 1024 * 1024); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		renderError(w, r, "Failed to parse upload. Please ensure the file is under 5MB.")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error reading form file: %v", err)
		renderError(w, r, "Failed to read uploaded file")
		return
	}
	defer file.Close()

	log.Printf("Uploaded file details - Filename: %s, Size: %d bytes", header.Filename, header.Size)

	// Check file size before processing
	if header.Size > 5*1024*1024 { // 5MB limit
		renderError(w, r, "Image size exceeds 5MB limit. Please choose a smaller image.")
		return
	}

	// Read the uploaded file content
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Error reading image content: %v", err)
		renderError(w, r, "Failed to process image")
		return
	}

	log.Println("Successfully read uploaded image content")

	// Encode the image content to base64
	encodedImage := base64.StdEncoding.EncodeToString(fileBytes)
	log.Println("Successfully encoded image to base64")

	// Call appropriate API to generate alt text
	altText, err := generateAltTextFunc(encodedImage)
	if err != nil {
		log.Printf("Error generating alt text: %v", err)
		renderError(w, r, formatErrorMessage(err.Error()))
		return
	}

	log.Printf("Generated alt text: %s", altText)

	// Process the alt text into options
	options := processAltText(altText)

	// Render the results
	component := components.AltTextResults(options)
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering results: %v", err)
		renderError(w, r, "Failed to display results")
	}
}

func renderError(w http.ResponseWriter, r *http.Request, message string) {
	component := components.UploadError(message)
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering error message: %v", err)
		http.Error(w, message, http.StatusInternalServerError)
	}
}

func processAltText(altText string) []string {
	options := strings.Split(altText, "\n")
	var cleanOptions []string

	for _, option := range options {
		option = strings.TrimSpace(option)
		if option != "" {
			// Remove numbering if present
			if strings.Contains("123456789", string(option[0])) && len(option) > 2 && option[1] == '.' {
				option = strings.TrimSpace(option[2:])
			}
			cleanOptions = append(cleanOptions, option)
		}
	}

	return cleanOptions
}

func formatErrorMessage(errMsg string) string {
	if strings.Contains(errMsg, "image exceeds 5 MB maximum") {
		return "Image size exceeds the 5MB limit. Please choose a smaller image."
	}
	if strings.Contains(errMsg, "invalid_request_error") {
		return "Invalid request. Please check your image and try again."
	}
	return "Failed to generate alt text. Please try again."
}
