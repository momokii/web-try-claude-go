package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

// OPEN AI DOCS api Reference
type OpenAI interface {
	OpenAISendMessage(content []OAMessageReq, with_format_response bool, format_response map[string]interface{}) (*OAChatCompletionResp, error)
	OpenAIGetFirstContentDataResp(content []OAMessageReq, with_format_response bool, format_response map[string]interface{}) (*OAMessage, error)
}

// Config holds the configuration for OpenAI API client
type Config struct {
	httpClient    *http.Client
	openAIBaseUrl string
	openAIModel   string
}

// default configuration for OpenAI API client
func DefaultConfig() *Config {
	return &Config{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		// user base url for chat completions endpoint with using gpt-4o-mini model
		openAIBaseUrl: "https://api.openai.com/v1/chat/completions",
		openAIModel:   "gpt-4o-mini",
	}
}

// client implementation for OpenAI API interfaces
type openaiAPI struct {
	apiKey             string
	openaiOrganization string
	openaiProject      string
	config             *Config
}

// client options for configuring the OpenAI API client
type ClientOption func(*Config)

func New(apiKey string, openaiOrganization string, openaiProject string, opts ...ClientOption) (OpenAI, error) {
	// from openai docs on
	// https://platform.openai.com/docs/api-reference/authentication
	// organization and project id is optional
	if apiKey == "" {
		return nil, errors.New("API Key is empty")
	}

	// create new OpenAI instance from private struct
	config := DefaultConfig()

	// apply custom options
	for _, opt := range opts {
		opt(config)
	}

	return &openaiAPI{
		apiKey:             apiKey,
		openaiOrganization: openaiOrganization,
		openaiProject:      openaiProject,
		config:             config,
	}, nil
}

// custom http client setup
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Config) {
		c.httpClient = httpClient
	}
}

// custom base url setup if need using different endpoint maybe like dalle or whisper or other
func WithBaseUrl(baseUrl string) ClientOption {
	return func(c *Config) {
		c.openAIBaseUrl = baseUrl
	}
}

// custom model setup if need using different model maybe like gpt-4o or gpt-4o-turbo or other
func WithModel(model string) ClientOption {
	return func(c *Config) {
		c.openAIModel = model
	}
}

// create response format for using JSON Schema for openai response format data request
// use this function to create response format for parameter in OpenAISendMessage()
// use with give the json schema name and json schema data
func OACreateResponseFormat(jsonName string, jsonSchema map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name": jsonName,
			"schema": map[string]interface{}{
				"type":       "object",
				"properties": jsonSchema,
			},
		},
	}
}

// base format response for request body parameter can get form OACreateResponseFormat()
// if need to response format, set the response format to the request body using OACreateResponseFormat()
func (c *openaiAPI) OpenAISendMessage(content []OAMessageReq, with_format_response bool, format_response map[string]interface{}) (*OAChatCompletionResp, error) {
	if c.apiKey == "" {
		return nil, errors.New("API Key is empty")
	}

	if with_format_response && format_response == nil {
		return nil, errors.New("format_response must be provided when with_format_response is true")
	}

	reqBody := OAReqBodyMessageCompletion{
		Model:    c.config.openAIModel,
		Messages: content,
	}

	// if using format response add response format to request body
	if with_format_response {
		reqBody.ResponseFormat = format_response
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.New("Failed to marshal request body")
	}

	// send req to openai
	req, err := http.NewRequest(http.MethodPost, c.config.openAIBaseUrl, bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		return nil, errors.New("Failed to create request")
	}

	// header setup
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := c.config.httpClient

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Failed to send request: " + err.Error())
	}
	defer func() {
		if resp.StatusCode != http.StatusOK {
			io.ReadAll(resp.Body)
		}
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Failed to send request: " + resp.Status)
	}

	// decode response
	var result OAChatCompletionResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.New("Failed to decode response: " + err.Error())
	}

	return &result, nil // return response

}

func (c *openaiAPI) OpenAIGetFirstContentDataResp(content []OAMessageReq, with_format_response bool, format_response map[string]interface{}) (*OAMessage, error) {
	// send request to openai
	resp, err := c.OpenAISendMessage(content, with_format_response, format_response)
	if err != nil {
		return nil, err
	}

	// get content first data
	data := resp.Choices[0].Message

	return &data, nil
}
