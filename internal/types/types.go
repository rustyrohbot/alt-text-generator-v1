package types

// TemplateData represents the data passed to HTML templates
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
