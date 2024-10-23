package claude

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type ClaudeAPI interface {
	ClaudeSendMessage(content []ClaudeMessageReq, maxToken int) (*ClaudeResp, error)
	ClaudeGetFirstContentDataResp(prompt []ClaudeMessageReq, maxToken int) (*ClaudeContentResp, error)
}

// Config holds the configuration for Claude API client
type Config struct {
	httpClient             *http.Client
	claudeBaseUrl          string
	claudeModel            string
	claudeAnthropicVersion string
}

// default configuration for Claude API client
func DefaultConfig() *Config {
	return &Config{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		// claude base url using the /messages endpoint default because this the only available endpoint on claude for now that can be user (text and vision)
		claudeBaseUrl:          "https://api.anthropic.com/v1/messages",
		claudeModel:            "claude-3-5-sonnet-20240620",
		claudeAnthropicVersion: "2021-06-01",
	}
}

// client implementation for Claude API interfaces
type claudeAPI struct {
	apiKey string
	config *Config
}

// client options for configuring the Claude API client
type ClientOption func(*Config)

func New(apiKey string, opts ...ClientOption) (ClaudeAPI, error) {

	if apiKey == "" {
		return nil, errors.New("API Key is empty")
	}

	// create new Claude API instance from private struct
	config := DefaultConfig()

	// apply options
	for _, opt := range opts {
		opt(config)
	}

	return &claudeAPI{
		apiKey: apiKey,
		config: config,
	}, nil
}

// custom options for configuring the Claude API client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Config) {
		c.httpClient = httpClient
	}
}

// custom options for configuring the Claude API client
func WithBaseUrl(baseUrl string) ClientOption {
	return func(c *Config) {
		c.claudeBaseUrl = baseUrl
	}
}

// custom options for configuring the Claude API client
func WithModel(model string) ClientOption {
	return func(c *Config) {
		c.claudeModel = model
	}
}

// custom options for configuring the Claude API client
func WithAnthropicVersion(version string) ClientOption {
	return func(c *Config) {
		c.claudeAnthropicVersion = version
	}
}

// the struct response will the same structure like the Claude API response form the docs
// Claude Response structure example by the Anthropic docs:
//
//	{
//		"id": "msg_01EcyWo6m4hyW8KHs2y2pei5",
//		"type": "message",
//		"role": "assistant",
//		"content": [
//		  {
//			"type": "text",
//			"text": "This image shows an ant, specifically a close-up view of an ant. The ant is shown in detail, with its distinct head, antennae, and legs clearly visible. The image is focused on capturing the intricate details and features of the ant, likely taken with a macro lens to get an extreme close-up perspective."
//		  }
//		],
//		"model": "claude-3-5-sonnet-20240620",
//		"stop_reason": "end_turn",
//		"stop_sequence": null,
//		"usage": {
//		  "input_tokens": 1551,
//		  "output_tokens": 71
//		}
//	  }
func (c *claudeAPI) ClaudeSendMessage(content []ClaudeMessageReq, maxToken int) (*ClaudeResp, error) {
	apiKey := c.apiKey
	if apiKey == "" {
		return nil, errors.New("API Key is empty")
	}

	reqBody := ClaudeReqBody{
		Model:     c.config.claudeModel,
		MaxTokens: maxToken,
		Message:   content,
	}

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.New("request failed: " + err.Error())
	}

	// send request to Claude
	req, err := http.NewRequest(http.MethodPost, c.config.claudeBaseUrl, bytes.NewBuffer(reqBodyJson))
	if err != nil {
		return nil, errors.New("request failed: " + err.Error())
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", c.config.claudeAnthropicVersion)
	req.Header.Set("Content-Type", "application/json")

	client := c.config.httpClient

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("request failed: " + err.Error())
	}
	defer func() {
		if resp.StatusCode != http.StatusOK {
			io.ReadAll(resp.Body)
		}
		resp.Body.Close()
	}()

	// error handling status
	if resp.StatusCode != http.StatusOK {
		var errClaude ClaudeRespError
		if err := json.NewDecoder(resp.Body).Decode(&errClaude); err != nil {
			return nil, errors.New("request failed with status code: " + resp.Status)
		}

		return nil, errors.New("Claude API response error: " + resp.Status + " with message: " + errClaude.Error.Message + " type: " + errClaude.Error.Type)
	}

	// decode response from Claude to map
	var result ClaudeResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.New("request failed: " + err.Error())
	}

	return &result, nil
}

// this can use if you want to get the first message content only (like not using conversational window and just doing "one shot" request)
// Request to Claude to get content data and process the response to get the first message content with return content structure as interface will like below example:
//
//	  {
//		"type": "text",
//		"text": "This image shows an ant, specifically a close-up view of an ant. The ant is shown in detail, with its distinct head, antennae, and legs clearly visible. The image is focused on capturing the intricate details and features of the ant, likely taken with a macro lens to get an extreme close-up perspective."
//	  }
func (c *claudeAPI) ClaudeGetFirstContentDataResp(prompt []ClaudeMessageReq, maxToken int) (*ClaudeContentResp, error) {
	// send request to Claude
	claudeResp, err := c.ClaudeSendMessage(prompt, maxToken)
	if err != nil {
		return nil, err
	}

	// with response example above
	// get content key from map and type assert as array of interface
	// get first element from array of interface and type assert as map
	content := claudeResp.Content[0]

	// return the message content as interface
	return &content, nil
}
