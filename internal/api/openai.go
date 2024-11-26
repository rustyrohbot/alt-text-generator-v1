package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const chatgptAPIURL = "https://api.openai.com/v1/completions"

func GenerateAltTextOpenAI(encodedImage string) (string, error) {
	log.Println("Reading OpenAI API key from environment variables")
	openaiAPIKey := os.Getenv("OPEN_AI_API_KEY")
	if openaiAPIKey == "" {
		log.Println("OpenAI API key is not set in environment variables")
		return "", fmt.Errorf("OpenAI API key is not set in environment variables")
	}
	log.Println("Successfully read OpenAI API key")

	prompt := `Generate 3 different alt text descriptions for this image. Vary the level of detail and focus in each description.
Each alt text should:
1. Be clear and concise
2. Avoid starting with "An image of" or "A photo of"
3. Focus on the most important elements
4. Use natural language

Return the descriptions in this format:
1. [first description]
2. [second description]
3. [third description]

Here's the base64 encoded image: %s`

	data := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": fmt.Sprintf(prompt, encodedImage)},
		},
		"max_tokens": 300,
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

	var chatResp struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}
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
