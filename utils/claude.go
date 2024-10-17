package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"scrapper-test/models"
)

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
func ClaudeSendMessage(content []models.ClaudeMessageReq) (interface{}, error) {
	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		return "", errors.New("API Key is empty")
	}

	reqBody := models.ClaudeReqBody{
		Model:     os.Getenv("CLAUDE_MODEL"),
		MaxTokens: 512 * 10,
		Message:   content,
	}

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// send request to Claude
	req, err := http.NewRequest("POST", os.Getenv("CLAUDE_BASE_URL"), bytes.NewBuffer(reqBodyJson))
	if err != nil {
		return "", err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", os.Getenv("CLAUDE_ANTHROPIC_VERSION"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// decode response from Claude to map
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result, nil
}

// Request to Claude to get content data and process the response to get the first message content with return content structure as interface will like below example:
//
//	  {
//		"type": "text",
//		"text": "This image shows an ant, specifically a close-up view of an ant. The ant is shown in detail, with its distinct head, antennae, and legs clearly visible. The image is focused on capturing the intricate details and features of the ant, likely taken with a macro lens to get an extreme close-up perspective."
//	  }
func ClaudeGetContentDataResp(prompt []models.ClaudeMessageReq) (map[string]interface{}, error) {
	// send request to Claude
	claudeResp, err := ClaudeSendMessage(prompt)
	if err != nil {
		return nil, err
	}

	// with response example above,
	// 1. type assert as map (JSON)
	// 2, get content key from map and type assert as array of interface
	claudeRespMsg, ok := claudeResp.(map[string]interface{})["content"].([]interface{})
	if !ok {
		return nil, errors.New("claude response not found")
	}

	// get first element from array of interface and type assert as map
	msg, ok := claudeRespMsg[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("claude response not found")
	}

	// return the message content as interface
	return msg, nil
}
