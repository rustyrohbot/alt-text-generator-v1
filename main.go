package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

const (
	chatgptAPIURL = "https://api.openai.com/v1/completions"
	claudeAPIURL  = "https://api.anthropic.com/v1/messages"
)

type TemplateData struct {
	Mode          string
	APIKeyMissing bool
}

// ChatGPTResponse represents the response from OpenAI API
type ChatGPTResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// ClaudeResponse represents the response from Anthropic API
type ClaudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

var tmpl = template.Must(template.ParseFiles("template.html"))

func main() {
	// Define flags for selecting which API to use
	useOpenAI := flag.Bool("openai", false, "Use OpenAI API")
	useAnthropic := flag.Bool("anthropic", false, "Use Anthropic API")
	flag.Parse()

	// Load environment variables from .env file
	log.Println("Loading environment variables from .env file")
	err := loadEnvFile(".env")
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file: %v", err)
	}
	log.Println("Successfully loaded .env file")

	// Set the appropriate API selection function and mode
	var generateAltTextFunc func(string) (string, error)
	var mode string
	if *useOpenAI {
		generateAltTextFunc = generateAltTextOpenAI
		mode = "openai"
	} else if *useAnthropic {
		generateAltTextFunc = generateAltTextClaude
		mode = "anthropic"
	} else {
		log.Fatalf("You must specify either -openai or -anthropic flag.")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		homeHandler(w, r, mode)
	})
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		uploadHandler(w, r, generateAltTextFunc, mode)
	})
	http.HandleFunc("/saveApiKey", saveApiKeyHandler)

	fmt.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func saveApiKeyHandler(w http.ResponseWriter, r *http.Request) {
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

	// Create or append to .env file
	f, err := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Failed to open .env file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// Check if the key already exists in the file
	if err := loadEnvFile(".env"); err == nil {
		existingKey := os.Getenv(envKey)
		if existingKey != "" {
			// Key exists, update it instead of appending
			updateEnvFile(".env", envKey, apiKey)
		} else {
			// Key doesn't exist, append it
			if _, err := f.WriteString(fmt.Sprintf("%s=%s\n", envKey, apiKey)); err != nil {
				http.Error(w, "Failed to write to .env file", http.StatusInternalServerError)
				return
			}
		}
	}

	// Set the environment variable for immediate use
	os.Setenv(envKey, apiKey)

	// Redirect back to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func updateEnvFile(filename, key, newValue string) error {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			lines[i] = fmt.Sprintf("%s=%s", key, newValue)
			break
		}
	}
	output := strings.Join(lines, "\n")
	return ioutil.WriteFile(filename, []byte(output), 0644)
}

func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			// Skip empty lines and comments
			if err == io.EOF {
				break
			}
			continue
		}

		// Split by the first '=' character to separate key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Invalid line, skip
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if err := os.Setenv(key, value); err != nil {
			return err
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

func homeHandler(w http.ResponseWriter, r *http.Request, mode string) {
	log.Println("Serving home page")

	// Check if API key exists
	var apiKey string
	if mode == "openai" {
		apiKey = os.Getenv("OPEN_AI_API_KEY")
	} else {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	data := TemplateData{
		Mode:          mode,
		APIKeyMissing: apiKey == "",
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request, generateAltTextFunc func(string) (string, error), mode string) {
	log.Println("Received upload request")
	if r.Method != http.MethodPost {
		log.Println("Invalid request method. Expected POST.")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
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
		http.Error(w, "API key not configured", http.StatusUnauthorized)
		return
	}

	contentType := r.Header.Get("Content-Type")
	log.Printf("Request Content-Type: %s", contentType)
	if contentType != "multipart/form-data" && !hasMultipartPrefix(contentType) {
		log.Println("Request Content-Type isn't multipart/form-data")
		http.Error(w, "Failed to read image file: Content-Type isn't multipart/form-data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error reading form file: %v", err)
		http.Error(w, "Failed to read image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Uploaded file details - Filename: %s, Size: %d bytes, Header: %v", header.Filename, header.Size, header.Header)

	// Read the uploaded file content
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Error reading image content: %v", err)
		http.Error(w, "Failed to read image content", http.StatusInternalServerError)
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
		http.Error(w, "Failed to generate alt text", http.StatusInternalServerError)
		return
	}

	log.Printf("Generated alt text: %s", altText)

	// Return alt text as response
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<div id='alt-text'>Generated Alt Text: %s</div>", altText)
	fmt.Fprintf(w, "<button hx-get='/'>Upload New Image</button>")
}

func hasMultipartPrefix(contentType string) bool {
	return len(contentType) >= 19 && contentType[:19] == "multipart/form-data"
}

func generateAltTextOpenAI(encodedImage string) (string, error) {
	log.Println("Reading OpenAI API key from environment variables")
	openaiAPIKey := os.Getenv("OPEN_AI_API_KEY")
	if openaiAPIKey == "" {
		log.Println("OpenAI API key is not set in environment variables")
		return "", fmt.Errorf("OpenAI API key is not set in environment variables")
	}
	log.Println("Successfully read OpenAI API key")

	prompt := fmt.Sprintf("Generate an alt text description for the following image encoded in base64: %s", encodedImage)
	log.Printf("Generated prompt for OpenAI: %s", prompt)

	data := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 100,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling JSON data: %v", err)
		return "", err
	}
	log.Println("Successfully marshaled request data to JSON")

	req, err := http.NewRequest("POST", chatgptAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)

	log.Println("Sending request to OpenAI API")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to OpenAI API: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Println("Successfully received response from OpenAI API")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", err
	}

	log.Printf("Response body: %s", body)

	var chatResp ChatGPTResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		log.Printf("Error unmarshaling response JSON: %v", err)
		return "", err
	}

	if len(chatResp.Choices) > 0 {
		log.Println("Successfully extracted response choice from ChatGPT")
		return chatResp.Choices[0].Text, nil
	}
	log.Println("No response choices from ChatGPT")
	return "", fmt.Errorf("No response from ChatGPT")
}

func generateAltTextClaude(encodedImage string) (string, error) {
	log.Println("Reading Anthropic API key from environment variables")
	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicAPIKey == "" {
		log.Println("Anthropic API key is not set in environment variables")
		return "", fmt.Errorf("Anthropic API key is not set in environment variables")
	}
	log.Println("Successfully read Anthropic API key")

	// Decode base64 image to get media type
	imageData, err := base64.StdEncoding.DecodeString(encodedImage)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %v", err)
	}

	// Create the request body with the correct structure for images
	data := map[string]interface{}{
		"model": "claude-3-opus-20240229",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Please generate a clear and concise alt text description for this image.",
					},
					{
						"type": "image",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": http.DetectContentType(imageData),
							"data":       encodedImage,
						},
					},
				},
			},
		},
		"max_tokens": 100,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling JSON data: %v", err)
		return "", err
	}
	log.Println("Successfully marshaled request data to JSON")

	req, err := http.NewRequest("POST", claudeAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", anthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	log.Println("Sending request to Anthropic API")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to Anthropic API: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Println("Successfully received response from Anthropic API")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", err
	}

	log.Printf("Response body: %s", body)

	// If we received an error response, parse and return it
	if strings.Contains(string(body), "error") {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return "", fmt.Errorf("API error: %s", errorResp.Error.Message)
		}
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		log.Printf("Error unmarshaling response JSON: %v", err)
		return "", err
	}

	if len(claudeResp.Content) > 0 {
		log.Println("Successfully extracted response from Claude")
		return claudeResp.Content[0].Text, nil
	}
	log.Println("No response from Claude")
	return "", fmt.Errorf("No response from Claude")
}
