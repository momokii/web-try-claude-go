package models

type ClaudeMessageReq struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeReqBody struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Message   []ClaudeMessageReq `json:"messages"`
}
