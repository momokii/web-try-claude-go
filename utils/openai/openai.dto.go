package openai

// OPEN AI DOCS api Reference
// https://platform.openai.com/docs/api-reference/chat/create

// OpenAIReqBodyMessageCompletion create document for openai dto
// if omitempty is used, the field will be omitted from the JSON representation of the object if the field has an empty value
// the omit empty value is optional in openai docs
type OAReqBodyMessageCompletion struct {
	Messages         interface{}            `json:"messages"` // required
	Model            string                 `json:"model"`    // required
	Store            bool                   `json:"store,omitempty"`
	Metadata         interface{}            `json:"metadata,omitempty"`
	FrequencyPenalty float64                `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]interface{} `json:"logit_bias,omitempty"`
	Logprobe         bool                   `json:"logprobe,omitempty"`
	Modalities       []string               `json:"modalities,omitempty"`
	ResponseFormat   map[string]interface{} `json:"response_format,omitempty"`
}

type OAMessageReq struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type OAContentVisionImageUrl struct {
	Url string `json:"url"`
}

type OAContentVisionBaseReq struct {
	Type     string                   `json:"type"`
	Text     *string                  `json:"text,omitempty"`
	ImageUrl *OAContentVisionImageUrl `json:"image_url,omitempty"`
}

// response OpenAI structure
type OAChatCompletionResp struct {
	ID                string     `json:"id"`
	Object            string     `json:"object"`
	Created           int64      `json:"created"`
	Model             string     `json:"model"`
	SystemFingerprint string     `json:"system_fingerprint"`
	Choices           []OAChoice `json:"choices"`
	Usage             OAUsage    `json:"usage"`
}

type OAChoice struct {
	Index        int       `json:"index"`
	Message      OAMessage `json:"message"`
	Logprobs     *string   `json:"logprobs"` // Could be null, so pointer
	FinishReason string    `json:"finish_reason"`
}

type OAMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OAUsage struct {
	PromptTokens           int          `json:"prompt_tokens"`
	CompletionTokens       int          `json:"completion_tokens"`
	TotalTokens            int          `json:"total_tokens"`
	CompletionTokensDetail TokensDetail `json:"completion_tokens_details"`
}

type TokensDetail struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}
