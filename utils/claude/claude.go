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
	ClaudeSendMessage(content *[]ClaudeMessageReq, maxToken int, with_custom_reqbody bool, req_body_custom *ClaudeReqBody) (*ClaudeResp, error)
	ClaudeGetFirstContentDataResp(prompt *[]ClaudeMessageReq, maxToken int, with_custom_reqbody bool, req_body_custom *ClaudeReqBody) (*ClaudeContentResp, error)
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

// New initializes a new ClaudeAPI client instance.
//
// This function is responsible for creating and configuring a new instance of the ClaudeAPI client.
// The `apiKey` parameter is required to authenticate requests to the Claude API, and optional client configurations
// can be provided through the `ClientOption` variadic parameters.
//
// Parameters:
//   - apiKey: A string containing the API key to authenticate with the Claude API. This is a required parameter.
//     If the API key is empty, the function will return an error.
//   - opts: A variadic list of `ClientOption` functions that allow custom configurations of the client. These options
//     can be used to customize things like base URL, timeouts, or the HTTP client used.
//
// Returns:
//   - ClaudeAPI: An interface representing the Claude API client. This interface is used to interact with the Claude API,
//     sending requests and handling responses.
//   - error: An error is returned if there is an issue creating the ClaudeAPI instance, such as if the API key is empty.
//
// Default Configuration Values:
//   - httpClient: The HTTP client used for making requests. By default, the client has a
//     timeout of 60 seconds (`http.Client{ Timeout: 60 * time.Second }`).
//   - claudeBaseUrl: The default base URL for the Claude API is set to the `/messages` endpoint,
//     as this is currently the primary endpoint available for both text and vision requests.
//     The default value is `"https://api.anthropic.com/v1/messages"`.
//   - claudeModel: The default model for message processing is `"claude-3-5-sonnet-20240620"`, which specifies
//     the Claude model version that will be used to generate responses.
//   - claudeAnthropicVersion: The API version used for interacting with Claude. The default value is `"2021-06-01"`.
//
// Example usage:
//
//	// Initialize ClaudeAPI with an API key
//	claudeClient, err := New("your-api-key")
//	if err != nil {
//	    log.Fatalf("Failed to create ClaudeAPI client: %v", err)
//	}
//
//	// Example with optional configurations
//	claudeClientWithOpts, err := New("your-api-key", WithTimeout(30*time.Second))
//	if err != nil {
//	    log.Fatalf("Failed to create ClaudeAPI client with options: %v", err)
//	}
//
// Function Details:
//
//  1. **API Key Check**: The `apiKey` parameter is mandatory, and the function checks whether it is provided.
//     If the `apiKey` is an empty string, the function returns an error (`"API Key is empty"`).
//  2. **Default Configuration**: A new configuration object is created using the `DefaultConfig` function,
//     which sets up standard settings like API endpoint, HTTP client, and other defaults.
//  3. **Client Options Application**: The function accepts a variadic list of `ClientOption` parameters. Each option
//     is applied to the configuration using a loop. `ClientOption` functions allow users to customize the client,
//     such as adjusting timeouts, changing the base URL, or configuring request headers.
//  4. **Client Creation**: A new instance of the private `claudeAPI` struct is created, containing the provided API key
//     and the configured options. This instance is returned as a `ClaudeAPI` interface.
//
// Notes:
//   - The `ClientOption` pattern provides flexibility in configuring the client without changing the core logic.
//   - Common custom options may include setting custom headers, timeout configurations, or API versioning.
//   - The returned `ClaudeAPI` instance can be used to interact with various Claude API endpoints such as sending messages.
//
// Considerations:
//   - If the API key is invalid or omitted, the client will not be able to authenticate requests to the Claude API.
//   - Ensure that the provided API key is correct and that any optional configurations align with your usage requirements.
//   - The function allows for easy extension and customization by accepting `ClientOption` functions.
//
// References:
// Initial Setup Claude: https://docs.anthropic.com/en/docs/initial-setup
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

// custom options for configuring the Claude API client, use it on New function initiate
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Config) {
		c.httpClient = httpClient
	}
}

// custom options for configuring the Claude API client, use it on New function initiate
func WithBaseUrl(baseUrl string) ClientOption {
	return func(c *Config) {
		c.claudeBaseUrl = baseUrl
	}
}

// custom options for configuring the Claude API client, use it on New function initiate
func WithModel(model string) ClientOption {
	return func(c *Config) {
		c.claudeModel = model
	}
}

// custom options for configuring the Claude API client, use it on New function initiate
func WithAnthropicVersion(version string) ClientOption {
	return func(c *Config) {
		c.claudeAnthropicVersion = version
	}
}

// ClaudeCreateOneContentImageVisionBase64 generates a vision content payload for uploading a base64-encoded image
// along with an optional text description to the Claude API.
//
// This function is designed to create a list of `ClaudeVisionContentBase` structures that can be used to send both
// an image and optional text content to Claude. The image data must be base64-encoded, and the supported image types. This function designed to make user easier to create the vision content for one image and one text data in one payload.
// include JPEG, PNG, GIF, and WebP.
//
// Parameters:
//   - media_type (string): The MIME type of the image being uploaded. Supported types include:
//   - "image/jpeg"
//   - "image/png"
//   - "image/gif"
//   - "image/webp"
//     If an unsupported media type is provided, the function returns an error with a message indicating the supported types.
//   - encode_file_base64 (string): The base64-encoded string representation of the image file. This should be the image
//     data encoded into base64 format.
//   - text_content (string): An optional parameter. If provided, this string will be added as a text component to accompany
//     the image. If no text content is provided, only the image content will be included in the payload.
//
// Returns:
//
//	([]ClaudeVisionContentBase, error): The function returns a slice of `ClaudeVisionContentBase` structs containing the
//	image (and text, if provided). If any required parameter (such as `media_type` or `encode_file_base64`) is missing,
//	or if an unsupported media type is provided, an error is returned.
//
// Example usage:
//
//	// Base64-encoded image data (as a placeholder example)
//	base64Image := "ixxx"
//
//	// Call the function to generate the vision content with image and optional text
//	visionContent, err := ClaudeCreateOneContentImageVisionBase64("image/png", base64Image, "This is a sample image.")
//	if err != nil {
//	    log.Fatalf("Error generating vision content: %v", err)
//	}
//
// Function Logic:
//
//  1. **Input Validation**: The function first checks if either `media_type` or `encode_file_base64` is empty.
//     If either of these values is missing, the function returns an error indicating that these fields are required.
//
//  2. **Supported Media Types**: The function verifies that the provided `media_type` is one of the supported image types:
//     `"image/jpeg"` - `"image/png"`- `"image/gif"`- `"image/webp"`
//
//     If the media type does not match one of these values, an error is returned, indicating the valid media types.
//
//  3. **Image Content**: A `ClaudeVisionContentBase` struct is created with the `type` set to `"image"`, and the base64-encoded
//     image data is embedded in the `Source` field, which is a nested `ClaudeVisionSource` struct. This struct includes
//     the `media_type` and `data` (the base64-encoded image string).
//
//  4. **Optional Text Content**: If the `text_content` parameter is provided, another `ClaudeVisionContentBase` struct is
//     created with the `type` set to `"text"`, and the provided text is stored in the `text` field. This is then appended
//     to the vision content slice, allowing both the image and text to be sent in a single request.
//
//  5. **Return**: The function returns the slice of `ClaudeVisionContentBase` structs, which includes the image content (and
//     text content, if provided), ready to be used in a Claude vision request.
//
// Notes:
//   - The function currently only supports base64-encoded images because Claude also for now on vision just support image with base64 encode data payload. Make sure to convert the image to a base64-encoded string
//     before passing it to the function.
//   - This function is designed to create a single vision content payload containing one image and an optional text.
//     if need Additional images can see through the Official docs (links) on references to create the modification on yourself. The base struct data is ClaudeVisionContentBase and ClaudeVisionSource.
//   - This function leverages Claude's recently added support for base64 image uploads as part of their vision capabilities,
//     making it possible to send images directly in requests for vision-related tasks.
//
// Considerations:
//   - When uploading images, ensure that the base64-encoded string is valid and corresponds to the correct media type.
//   - Large image files may need to be appropriately sized or compressed before encoding them to base64, as this may impact
//     the request size or processing time.
//
// References:
//   - Official Claude API documentation: https://docs.anthropic.com/en/api/messages-examples
func ClaudeCreateOneContentImageVisionBase64(media_type string, encode_file_base64 string, text_content string) ([]ClaudeVisionContentBase, error) {

	if media_type == "" || encode_file_base64 == "" {
		return nil, errors.New("media type or encode file base64 is empty")
	}

	if media_type != "image/jpeg" && media_type != "image/png" && media_type != "image/gif" && media_type != "image/webp" {
		return nil, errors.New("media type not supported, supported type: image/jpeg, image/png, image/gif, and image/webp")
	}

	content := []ClaudeVisionContentBase{
		{
			Type: "image",
			Source: &ClaudeVisionSource{
				Type:      "base64",
				MediaType: media_type,
				Data:      encode_file_base64,
			},
		},
	}

	if text_content != "" {
		content = append(content, ClaudeVisionContentBase{
			Type: "text",
			Text: &text_content,
		})
	}

	return content, nil
}

// ClaudeSendMessage sends a message to the Claude API and returns the response.
//
// This function constructs and sends a request to Claude, either using a custom request body or
// building a default request body from the provided content and configuration.
// It handles the response, including error handling, and returns the parsed response.
//
// Parameters:
//   - content: A pointer to a slice of `ClaudeMessageReq` containing the messages to be sent to Claude.
//     Each message includes a `role` (e.g., "user", "system") and `content` which can be text or vision data.
//   - maxToken: An integer specifying the maximum number of tokens (words) allowed in the response (you can set to 0 if using custom_reqbody because the token itself you will provide inside the custom reqbody).
//   - with_custom_reqbody: A boolean flag indicating whether a custom request body should be used.
//   - req_body_custom: A pointer to `ClaudeReqBody`, which is used if `with_custom_reqbody` is true.
//     If this value is nil when `with_custom_reqbody` is true, an error is returned.
//
// Returns:
//   - A pointer to `ClaudeResp`, which contains the ID, content, model, and usage statistics of the response.
//   - An error if the request fails at any stage.
//
// Example usage:
//
//	// Define message content
//	messages := []ClaudeMessageReq{
//	    {Role: "user", Content: "What is the weather today?"},
//	}
//
//	// Send request with default body
//	response, err := claudeAPI.ClaudeSendMessage(&messages, 100, false, nil)
//	if err != nil {
//	    log.Fatalf("Failed to send message to Claude: %v", err)
//	}
//	fmt.Println("Claude response:", response)
//
//	// Send request with custom body
//	customReqBody := ClaudeReqBody{
//	    Model:     "claude-v1",
//	    MaxTokens: 150,
//	    Message:   messages,
//	}
//	response, err := claudeAPI.ClaudeSendMessage(nil, 0, true, &customReqBody)
//	if err != nil {
//	    log.Fatalf("Failed to send message to Claude with custom body: %v", err)
//	}
//	fmt.Println("Claude custom response:", response)
//
// Function Details:
//
//  1. API Key Validation: The function checks if the `apiKey` is empty, and returns an error if it is missing.
//  2. Custom Request Body: If `with_custom_reqbody` is true, the custom request body (`req_body_custom`) is used.
//     If it's nil, an error is returned.
//  3. Default Request Body: If `with_custom_reqbody` is false, the function creates a default `ClaudeReqBody` with
//     the specified model, max tokens, messages, and a temperature of 1.0.
//  4. Request Construction: The request is sent as a JSON payload to Claude's API endpoint using an HTTP POST method.
//  5. Headers: The request includes the necessary headers, such as the API key, version, and content type.
//  6. Response Handling: If the response status code is not 200 (OK), the function decodes the error response from Claude
//     and returns a detailed error message. The successful response is decoded into a `ClaudeResp` struct.
//  7. Error Handling: The function provides clear error messages for request building, sending, and response decoding failures.
//
// Notes:
//   - The `content` field of each message can include text or image data for vision-based requests, using the `ClaudeContentVision` structure.
//   - The function uses the configured HTTP client (`c.config.httpClient`) to send the request.
//   - The `ClaudeResp` structure includes usage statistics (e.g., input and output tokens) and any applicable stop sequences.
//
// References:
//   - Official Claude API documentation: https://docs.anthropic.com/en/api/messages
func (c *claudeAPI) ClaudeSendMessage(content *[]ClaudeMessageReq, maxToken int, with_custom_reqbody bool, req_body_custom *ClaudeReqBody) (*ClaudeResp, error) {

	var reqBody interface{}

	apiKey := c.apiKey
	if apiKey == "" {
		return nil, errors.New("API Key is empty")
	}

	if with_custom_reqbody && req_body_custom == nil {
		return nil, errors.New("request failed: custom request body is empty")
	}

	if !with_custom_reqbody && content == nil {
		return nil, errors.New("request failed: content is empty")
	}

	if with_custom_reqbody {
		reqBody = req_body_custom

	} else {
		reqBody = ClaudeReqBody{
			Model:       c.config.claudeModel,
			MaxTokens:   maxToken,
			Messages:    *content,
			Temperature: 1.0, // default value from docs
		}
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

// ClaudeGetFirstContentDataResp sends a prompt to the Claude API and returns the first content response.
//
// Notes: --
// This function is designed to send a message to the Claude API using the provided prompt,
// retrieve the full response, and extract the first content element (normally is the text type with content is the answer from model) from the response that can use for simplicity reason if you just need to use it like just the content, so you can only the return content straight away and not the full response structure of Claude Response.
//
// Parameters:
//   - prompt: A pointer to a slice of `ClaudeMessageReq` containing the messages to be sent to Claude.
//     Each message includes a `role` (e.g., "user", "system") and `content` (text or vision data).
//   - maxToken: An integer specifying the maximum number of tokens (words) allowed in the response.
//   - with_custom_reqbody: A boolean flag indicating whether a custom request body should be used.
//   - req_body_custom: A pointer to `ClaudeReqBody`, which is used if `with_custom_reqbody` is true.
//     If this value is nil when `with_custom_reqbody` is true, an error is returned.
//
// Returns:
//   - A pointer to `ClaudeContentResp`, representing the first content element returned by Claude in the response.
//   - An error if the request fails or if there is an issue extracting the content.
//
// Example usage:
//
//	// Define prompt messages
//	messages := []ClaudeMessageReq{
//	    {Role: "user", Content: "Summarize the latest news."},
//	}
//
//	// Send request and get the first content response
//	firstContent, err := claudeAPI.ClaudeGetFirstContentDataResp(&messages, 100)
//	if err != nil {
//	    log.Fatalf("Failed to get first content data: %v", err)
//	}
//	fmt.Println("First content data:", firstContent.Text)
//
// Function Details:
//
//  1. **ClaudeSendMessage Call**: This function internally calls `ClaudeSendMessage` to send the provided prompt to Claude.
//     It uses the default request body (without custom modifications) and a specified maximum token limit.
//  2. **Response Parsing**: Once the response is returned by `ClaudeSendMessage`, the function extracts the `Content` field from the response.
//  3. **First Content Extraction**: The function retrieves the first element from the `Content` array of `ClaudeResp`.
//     If successful, this content is returned as `ClaudeContentResp`.
//  4. **Error Handling**: If there is any error in sending the request or parsing the response, the error is returned directly.
//
// Notes:
//   - This function simplifies the process of retrieving the first content element from a Claude API response.
//   - The `Content` field in the Claude response is an array, and this function assumes at least one element is present.
//     If the array is empty, this would result in an index error.
//   - The `ClaudeContentResp` structure contains the `Type` and `Text` fields representing the response content.
//
// Considerations:
//   - You may need to check for errors in the returned content, such as if the array is empty or the first element is invalid.
//   - This function is designed to handle textual responses, though the Claude API can also support other content types (e.g., vision data) that still you can pass image data here on base64 encoding with structure that Claude needs.
//
// References:
//   - Official Claude API documentation: https://docs.anthropic.com/en/api/messages
func (c *claudeAPI) ClaudeGetFirstContentDataResp(prompt *[]ClaudeMessageReq, maxToken int, with_custom_reqbody bool, req_body_custom *ClaudeReqBody) (*ClaudeContentResp, error) {
	// send request to Claude
	claudeResp, err := c.ClaudeSendMessage(prompt, maxToken, with_custom_reqbody, req_body_custom)
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
