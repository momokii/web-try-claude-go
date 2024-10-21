package claude

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type ClaudeAPI interface {
	ClaudeSendMessage(content []ClaudeMessageReq, maxToken int) (*ClaudeResp, error)
	ClaudeGetFirstContentDataResp(prompt []ClaudeMessageReq, maxToken int) (*ClaudeContentResp, error)
}

type claudeAPI struct {
	apiKey                 string
	claudeBaseUrl          string
	claudeModel            string
	claudeAnthropicVersion string
}

func New(apiKey string, claudeBaseUrl string, claudeModel string, claudeAnthropicVersion string) ClaudeAPI {
	return &claudeAPI{
		apiKey:                 apiKey,
		claudeBaseUrl:          claudeBaseUrl,
		claudeModel:            claudeModel,
		claudeAnthropicVersion: claudeAnthropicVersion,
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
		Model:     c.claudeModel,
		MaxTokens: maxToken,
		Message:   content,
	}

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// send request to Claude
	req, err := http.NewRequest("POST", c.claudeBaseUrl, bytes.NewBuffer(reqBodyJson))
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", c.claudeAnthropicVersion)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// decode response from Claude to map
	var result ClaudeResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
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
	msg := claudeResp.Content[0]

	// return the message content as interface
	return &msg, nil
}
