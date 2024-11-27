package handlers

import (
	"encoding/base64"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func UploadHandler(w http.ResponseWriter, r *http.Request, generateAltTextFunc func(string) (string, error), mode string) {
	log.Println("Received upload request")

	// Add debug logging
	log.Printf("Request Method: %s", r.Method)
	log.Printf("Content Type: %s", r.Header.Get("Content-Type"))

	if r.Method != http.MethodPost {
		log.Println("Invalid request method. Expected POST.")
		renderUploadError(w, "Invalid request method")
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
		renderUploadError(w, "API key not configured")
		return
	}

	// Try to parse the multipart form with a 6MB limit (slightly higher than our 5MB limit to account for form overhead)
	if err := r.ParseMultipartForm(6 * 1024 * 1024); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		renderUploadError(w, "Failed to parse upload. Please ensure the file is under 5MB.")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error reading form file: %v", err)
		renderUploadError(w, "Failed to read uploaded file. Please try again.")
		return
	}
	defer file.Close()

	log.Printf("Uploaded file details - Filename: %s, Size: %d bytes, Header: %v", header.Filename, header.Size, header.Header)

	// Check file size before processing
	if header.Size > 5*1024*1024 { // 5MB limit
		renderUploadError(w, "Image size exceeds 5MB limit. Please choose a smaller image.")
		return
	}

	// Read the uploaded file content
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Error reading image content: %v", err)
		renderUploadError(w, "Failed to process image")
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
		renderUploadError(w, formatErrorMessage(err.Error()))
		return
	}

	log.Printf("Generated alt text: %s", altText)

	// Return success response
	renderSuccess(w, altText)
}

func renderSuccess(w http.ResponseWriter, altText string) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
        <div class="bg-green-50 border border-green-400 text-green-700 px-4 py-3 rounded-lg">
            <h3 class="font-bold mb-4">Generated Alt Text Options:</h3>
            <div class="space-y-4">%s</div>
            <button onclick="location.reload()" class="mt-4 bg-green-100 text-green-700 px-4 py-2 rounded hover:bg-green-200">
                Upload New Image
            </button>
        </div>
    `, formatAltTextOptions(altText))
}

func formatAltTextOptions(altText string) string {
	options := strings.Split(altText, "\n")
	var formatted strings.Builder

	for _, option := range options {
		option = strings.TrimSpace(option)
		if option != "" {
			formatted.WriteString(fmt.Sprintf(`
                <div class="bg-white p-3 rounded border border-green-200">
                    <p>%s</p>
                </div>
            `, html.EscapeString(option)))
		}
	}

	return formatted.String()
}

func renderUploadError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<div class="bg-red-50 border border-red-400 text-red-700 px-4 py-3 rounded-lg">
			<p class="font-bold mb-2">Error: %s</p>
			<button 
				onclick="document.getElementById('uploadForm').reset(); this.closest('.bg-red-50').remove()"
				class="bg-red-100 text-red-700 px-4 py-2 rounded hover:bg-red-200"
			>
				Try Again
			</button>
		</div>
	`, message)
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

func hasMultipartPrefix(contentType string) bool {
	return len(contentType) >= 19 && contentType[:19] == "multipart/form-data"
}
