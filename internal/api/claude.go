package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const claudeAPIURL = "https://api.anthropic.com/v1/messages"

func GenerateAltTextClaude(encodedImage string) (string, error) {
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

	prompt := `Generate 3 different alt text descriptions for this image. Vary the level of detail and focus in each description.
Each alt text should:
1. Be clear and concise
2. Avoid starting with "An image of" or "A photo of"
3. Focus on the most important elements
4. Use natural language

Return the descriptions in this format:
1. [first description]
2. [second description]
3. [third description]`

	// Create the request body with the correct structure for images
	data := map[string]interface{}{
		"model": "claude-3-opus-20240229",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
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
		"max_tokens": 300,
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

	var claudeResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
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
