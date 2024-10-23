package claude

// create document for claude dto

type ClaudeMessageReq struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeReqBody struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Message   []ClaudeMessageReq `json:"messages"`
}

type ClaudeContentVision struct {
	Type   string `json:"type"`
	Source struct {
		Type      string `json:"type"`
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	} `json:"source"`
}

type ClaudeRespError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

type ClaudeContentResp struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeResp struct {
	ID           string              `json:"id"`
	Type         string              `json:"type"`
	Role         string              `json:"role"`
	Content      []ClaudeContentResp `json:"content"`
	Model        string              `json:"model"`
	StopReason   string              `json:"stop_reason"`
	StopSequence string              `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}
