package models

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeReqBody struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Message   []Message `json:"messages"`
}
