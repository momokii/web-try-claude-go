package claude

// message bidy content structure
type ClaudeMessageReq struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // if need using vision can send image to content
	// https://docs.anthropic.com/en/api/messages
}

// content structure for vision
type ClaudeVisionSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// Struct untuk data tipe image dan text
type ClaudeVisionContentBase struct {
	Type   string              `json:"type"`
	Source *ClaudeVisionSource `json:"source,omitempty"` // Using pointer to allow nil value
	Text   *string             `json:"text,omitempty"`   // using pointer to allow nil value
}

// claude full request body structure with all possible fields
type ClaudeReqBody struct {
	Model         string                   `json:"model"`      // required
	MaxTokens     int                      `json:"max_tokens"` // required
	Messages      []ClaudeMessageReq       `json:"messages"`   // required
	Metadata      map[string]interface{}   `json:"metadata,omitempty"`
	StopSequences []string                 `json:"stop_sequences,omitempty"`
	Stream        bool                     `json:"stream,omitempty"`
	System        string                   `json:"system,omitempty"`
	Temperature   float64                  `json:"temperature,omitempty"` // default 1.0
	ToolChoice    map[string]interface{}   `json:"tool_choice,omitempty"`
	Tools         []map[string]interface{} `json:"tools,omitempty"`
}

// Claude 4xx error response structure
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

// claude full response structure on chat completions
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
