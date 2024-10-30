package openai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

const (
	OAUrlBase                  = "https://api.openai.com/v1"
	OAUrlTextCompletions       = OAUrlBase + "/chat/completions"
	OAUrlImageGenerationsDallE = OAUrlBase + "/images/generations"
	OAUrlTextToSpeech          = OAUrlBase + "/audio/speech"
)

type OpenAI interface {
	OpenAISendMessage(content *[]OAMessageReq, with_format_response bool, format_response *map[string]interface{}, with_custom_reqbody bool, req_body_custom *OAReqBodyMessageCompletion) (*OAChatCompletionResp, error)
	OpenAIGetFirstContentDataResp(content *[]OAMessageReq, with_format_response bool, format_response *map[string]interface{}, with_custom_reqbody bool, req_body_custom *OAReqBodyMessageCompletion) (*OAMessage, error)
	OpenAICreateImageDallE(req_body *OAReqImageGeneratorDallE) (*OAImageGeneratorDallEResp, error)
	OpenAITextToSpeech(req_body *OAReqTextToSpeech) (*OATextToSpeechResp, error)
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
		openAIBaseUrl: OAUrlTextCompletions,
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

// New creates and returns a new instance of the OpenAI API client.
//
// This function initializes the OpenAI client with the provided API key, organization, and project ID.
// Additionally, it applies any optional configuration options passed via the variadic `opts` parameter.
//
// Parameters:
//   - apiKey: The API key used to authenticate with OpenAI's API. This is required and cannot be empty.
//   - openaiOrganization: (Optional) The organization ID associated with your OpenAI account. This can be omitted if not applicable.
//   - openaiProject: (Optional) The project ID associated with your OpenAI account. This can be omitted if not applicable.
//   - opts: A variadic parameter that accepts one or more `ClientOption` functions to customize the client's behavior.
//
// Returns:
//   - An OpenAI client instance, or an error if the API key is missing or another issue occurs during initialization.
//
// Default Configuration Values:
//   - httpClient: The HTTP client used for making requests. By default, the client has a
//     timeout of 60 seconds (`http.Client{ Timeout: 60 * time.Second }`).
//   - openAIBaseUrl: The default base URL for the OpenAI is set to the `/chat/completions` endpoint
//     The default value is `"https://api.openai.com/v1/chat/completions"`.
//   - openAIModel: The default model for message processing is `"gpt-4o-mini"`, which specifies
//     the Claude model version that will be used to generate responses.
//
// Example usage:
//
//	// Initialize the OpenAI client with an API key
//	openaiClient, err := New("your-api-key", "your-org-id", "your-project-id")
//	if err != nil {
//	    log.Fatalf("Failed to create OpenAI client: %v", err)
//	}
//
//	// Optionally, provide custom options such as timeouts, base URL, or HTTP clients
//	customClient, err := New("your-api-key", "your-org-id", "your-project-id", WithCustomTimeout(30 * time.Second))
//	if err != nil {
//	    log.Fatalf("Failed to create OpenAI client with custom options: %v", err)
//	}
//
// Notes:
//   - The `apiKey` is required and must be provided, otherwise an error will be returned.
//   - The `openaiOrganization` and `openaiProject` parameters are optional and can be left empty if not needed.
//   - `ClientOption` is a functional option pattern that allows customization of the client, such as setting custom HTTP clients or changing API base URLs.
//
// References:
//   - Official OpenAI API authentication: https://platform.openai.com/docs/api-reference/authentication
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

// use if need custom http client setup, use it on New function initiate
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Config) {
		c.httpClient = httpClient
	}
}

// custom base url setup if need using different endpoint maybe like dalle or whisper or other, use it on New function initiate
func WithBaseUrl(baseUrl string) ClientOption {
	return func(c *Config) {
		c.openAIBaseUrl = baseUrl
	}
}

// custom model setup if need using different model maybe like gpt-4o or gpt-4o-turbo or other, use it on New function initiate
func WithModel(model string) ClientOption {
	return func(c *Config) {
		c.openAIModel = model
	}
}

// OACreateResponseFormat creates a response format using a JSON Schema for OpenAI response format data requests.
//
// This function is used to generate a JSON Schema structure that can be passed as a parameter
// to the OpenAISendMessage() function, providing a standard format for the response.
//
// Parameters:
//   - jsonName: A string representing the name of the JSON schema.
//   - jsonSchema: A map of string to interface, representing the schema data, specifically the properties
//     of the schema as defined by the OpenAI structured output documentation.
//
// Returns:
//   - A map[string]interface{} representing the formatted response structure using JSON Schema,
//     including the schema name and its associated properties.
//
// Example usage:
//
//		jsonSchema := map[string]interface{}{
//	 "type": "object",
//		"properties": map[string]interface{}{
//			  "title": map[string]interface{}{
//			    "type": "string",
//			  },
//			  "description": map[string]interface{}{
//			    "type": "string",
//			  },
//		   },
//		}
//
//		formattedResponse := OACreateResponseFormat("MySchema", jsonSchema)
//		fmt.Printf("Formatted response: %v\n", formattedResponse)
//
// JSON Schema Structure:
//   - The structure returned by this function will conform to the schema guidelines provided by OpenAI.
//     More details and examples can be found at the following link:
//     https://platform.openai.com/docs/guides/structured-outputs/examples
//
// Returned Structure Example:
//
//	{
//	  "type": "json_schema",
//	  "json_schema": {
//	    "name": "MySchema",
//	    "schema": {
//	      "type": "object",
//	      "properties": {
//	        "title": {
//	          "type": "string"
//	        },
//	        "description": {
//	          "type": "string"
//	        }
//	      }
//	    }
//	  }
//	}
func OACreateResponseFormat(jsonName string, jsonSchema map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name":   jsonName,
			"schema": jsonSchema,
		},
	}
}

// OACreateOneContentVision constructs a vision content payload for uploading an image (either as a URL or base64-encoded string)
// along with optional text to the OpenAI API.
//
// This function creates a list of `OAContentVisionBaseReq` structures, enabling you to upload both image data (via URL or base64 encoding)
// and optional descriptive text to OpenAI's vision endpoint. Supported media types include JPEG, PNG, JPG, GIF, and WebP.
//
// Parameters:
//   - media_type (string): The MIME type of the image when using base64 encoding. This is required when `using_image_url` is false.
//     Supported types include:
//   - "image/png"
//   - "image/jpeg"
//   - "image/jpg"
//   - "image/gif"
//   - "image/webp"
//   - using_image_url (bool): Specifies whether the image is provided as a URL or a base64-encoded string.
//   - If `true`, the function expects `url_or_base64encoding` to be a valid URL.
//   - If `false`, the function expects `url_or_base64encoding` to be a base64-encoded image string, and `media_type` must be provided.
//   - url_or_base64encoding (string): The image data provided as either a URL (when `using_image_url` is `true`) or a base64-encoded
//     string (when `using_image_url` is `false`). If this value is empty, the function returns an error indicating that both
//     `media_type` and `url_or_base64encoding` must be provided.
//   - text_content (string): An optional text string to accompany the image content. This will be included as a separate text
//     component if provided.
//
// Returns:
//
//	([]OAContentVisionBaseReq, error): A slice of `OAContentVisionBaseReq` structs containing the image (and optional text).
//	If the required parameters are not met or the media type is unsupported, an error is returned.
//
// Example usage:
//
//	// Example URL-based request
//	visionContent, err := OACreateOneContentVision("", true, "https://example.com/sample-image.jpg", "This is an example image.")
//	if err != nil {
//	    log.Fatalf("Error generating vision content: %v", err)
//	}
//
//	// Example base64-encoded request
//	base64Image := "iVBORw0KGgoAAAANSUhEUgAAAQAAAAEACAYAAABccqhmAAAACXBIWXMAAB7CAAAewgFu0HU+AAAK..." // truncated for example
//	visionContent, err := OACreateOneContentVision("image/png", false, base64Image, "This is an example image.")
//	if err != nil {
//	    log.Fatalf("Error generating vision content: %v", err)
//	}
//
// Function Logic:
//  1. **Input Validation**: Checks if `url_or_base64encoding` is empty. If it is, returns an error requiring both
//     `media_type` and `url_or_base64encoding` to be provided. If using base64 encoding (`using_image_url` is false), `media_type`
//     must also be specified.
//  2. **Supported Media Types**: Validates that `media_type` is one of the supported image types if `using_image_url` is false.
//     Supported types include "image/png", "image/jpeg", "image/jpg", "image/gif", and "image/webp". If an unsupported type
//     is provided, the function returns an error listing the valid types.
//  3. **Data Preparation**: If `using_image_url` is true, `imageData` is set to `url_or_base64encoding`. Otherwise, `imageData`
//     is created as a "data URI" by prepending `media_type` and "base64," to the encoded string. This format complies with the
//     OpenAI API's expectations for base64-encoded images.
//  4. **Image Content**: A `OAContentVisionBaseReq` struct is created for the image content, setting `Type` to `"image_url"`,
//     and the image data is assigned to the `ImageUrl` field.
//  5. **Optional Text Content**: If `text_content` is provided, another `OAContentVisionBaseReq` struct is appended with `Type` set
//     to `"text"` and `Text` containing the provided text. This allows both image and text content to be sent in a single request.
//  6. **Return**: Returns the slice of `OAContentVisionBaseReq` structs, which includes the image content (and text content,
//     if provided), ready for an OpenAI vision request.
//
// Notes:
//   - Ensure that the `url_or_base64encoding` contains a valid URL when `using_image_url` is true or a base64-encoded image string
//     when false.
//   - This function supports only a single image and an optional text. Multiple images or additional text content would
//     require separate calls or modifications to the function.
//   - OpenAI’s API currently supports base64 and URL images as part of its vision feature, making it possible to use both methods
//     with this function.
//
// Considerations:
//   - Base64-encoded images should be appropriately sized or compressed before encoding to avoid excessively large requests.
//   - URLs provided should be publicly accessible or authenticated as needed by the OpenAI API.
//   - this function hope can make you easier for send vision content if just contain one image and optional text content, if you need more than one image, you can create your own structure based on OpenAI Docs with struct OAContentVisionBaseReq & OAContentVisionImageUrl (for content structure) and append it to the slice of OAContentVisionBaseReq
//
// Reference for Vision OpenAI Docs:
// - Official OpenAI API documentation: https://platform.openai.com/docs/guides/vision
func OACreateOneContentVision(media_type string, using_image_url bool, url_or_base64encoding string, text_content string) ([]OAContentVisionBaseReq, error) {
	if url_or_base64encoding == "" {
		return nil, errors.New("media_type and url_or_base64encoding must be provided")
	}

	if media_type == "" && !using_image_url {
		return nil, errors.New("media_type must be provided when using base64 encoding")
	}

	if !using_image_url && media_type != "image/png" && media_type != "image/jpeg" && media_type != "image/jpg" && media_type != "image/gif" && media_type != "image/webp" {
		return nil, errors.New("media_type must be image/png, image/jpeg, or image/jpg")
	}

	var imageData string

	// data url or base64 encoding and the format is based on OpenAI API Docs
	if using_image_url {
		imageData = url_or_base64encoding
	} else {
		imageData = "data:" + media_type + ";base64," + url_or_base64encoding
	}

	contentVision := []OAContentVisionBaseReq{
		{
			Type: "image_url",
			ImageUrl: &OAContentVisionImageUrl{
				Url: imageData,
			},
		},
	}

	if text_content != "" {
		contentVision = append(contentVision, OAContentVisionBaseReq{
			Type: "text",
			Text: &text_content,
		})
	}

	return contentVision, nil
}

// OpenAISendMessage sends a message to OpenAI's API and handles the request and response format.
//
// This function creates and sends a request to the OpenAI API, allowing for custom request bodies and response formats.
// It either uses a provided custom request body or constructs a request body based on the provided message content.
// If response formatting is required, the `OACreateResponseFormat()` function can be used to generate the response format schema.
//
// Parameters:
//   - content: A pointer to a slice of OAMessageReq, which represents the request message content to be sent to OpenAI.
//     This is used if `with_custom_reqbody` is set to false.
//   - with_format_response: A boolean indicating whether a response format should be applied. If true, `format_response` must be provided.
//   - format_response: A map containing the JSON schema for formatting the response (can be created using OACreateResponseFormat).
//   - with_custom_reqbody: A boolean indicating whether a custom request body (`req_body_custom`) should be used.
//   - req_body_custom: A pointer to an OAReqBodyMessageCompletion struct. This is used if `with_custom_reqbody` is true.
//
// Returns:
//   - A pointer to an OAChatCompletionResp struct containing the API response.
//   - An error if the request fails, or if invalid parameters are provided.
//
// Example usage:
//
//	content := []OAMessageReq{
//	  {Role: "user", Content: "What is the weather like today?"},
//	}
//
//	formatResponse := OACreateResponseFormat("WeatherResponse", map[string]interface{}{
//	  "temperature": map[string]interface{}{"type": "string"},
//	  "condition": map[string]interface{}{"type": "string"},
//	})
//
//	response, err := openaiAPIInstance.OpenAISendMessage(&content, true, formatResponse, false, nil)
//	if err != nil {
//	    log.Fatalf("Failed to send message: %v", err)
//	}
//	fmt.Printf("API response: %+v\n", response)
//
// Notes:
//   - The function checks for invalid states, such as missing content or custom request bodies when required.
//   - The request is sent as a POST request with a JSON payload, and the response is decoded into the OAChatCompletionResp struct.
//
// References:
// - Official OpenAI API documentation: https://platform.openai.com/docs/api-reference/chat/create
func (c *openaiAPI) OpenAISendMessage(content *[]OAMessageReq, with_format_response bool, format_response *map[string]interface{}, with_custom_reqbody bool, req_body_custom *OAReqBodyMessageCompletion) (*OAChatCompletionResp, error) {

	// var reqBody interface{}
	var reqBody interface{}

	if c.apiKey == "" {
		return nil, errors.New("API Key is empty")
	}

	// check if with_format_response is true, format_response must be provided
	if with_format_response && format_response == nil {
		return nil, errors.New("format_response must be provided when with_format_response is true")
	}

	// check if with_custom_reqbody is true, req_body_custom must be provided
	if with_custom_reqbody && req_body_custom.Messages == nil {
		return nil, errors.New("req_body_custom must be provided when with_custom_reqbody is true")
	}

	// check if with_custom_reqbody is false, content must be provided
	if !with_custom_reqbody && content == nil {
		return nil, errors.New("content must be provided")
	}

	// create request body
	if with_custom_reqbody {

		if with_format_response {
			req_body_custom.ResponseFormat = *format_response
		}

		reqBody = req_body_custom

	} else {
		reqData := OAReqBodyMessageCompletion{
			Model:    c.config.openAIModel,
			Messages: content,
		}

		// if using format response add response format to request body
		if with_format_response {
			reqData.ResponseFormat = *format_response
		}

		reqBody = reqData
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

// OpenAIGetFirstContentDataResp retrieves the first content data from an OpenAI API response.
//
// This function sends a message request to the OpenAI API using the given content,
// and then extracts the first response content that basically the message response message from the API's response that can use for simplicity reason if you just need to use it like "one shot" request, so you can only the return content straight away and not the full response structure of OpenAI Response.
//
// Parameters:
//   - content: A pointer to a slice of OAMessageReq, which represents the request message content to be sent to OpenAI.
//   - with_format_response: A boolean indicating whether the response should be formatted.
//   - format_response: A map that contains additional formatting options for the response. if you need to use the format_response that supported by OpenAI API. Official Docs and structure about structured response OpenAPI schema in: https://platform.openai.com/docs/guides/structured-outputs/examples
//
// Returns:
//   - A pointer to an OAMessage struct that contains the first content data from the response.
//   - An error if the request to OpenAI fails.
//
// Example usage:
//
//	content := []OAMessageReq{...}
//	formatOptions := map[string]interface{}{
//	  "option1": "value1",
//	  // add formatting options here
//	}
//	firstContent, err := openaiAPIInstance.OpenAIGetFirstContentDataResp(&content, true, formatOptions)
//	if err != nil {
//	    log.Fatalf("Failed to get first content data: %v", err)
//	}
//	fmt.Println("First response content:", firstContent)
//
// References:
// - Official OpenAI API documentation: https://platform.openai.com/docs/api-reference/chat/create
func (c *openaiAPI) OpenAIGetFirstContentDataResp(content *[]OAMessageReq, with_format_response bool, format_response *map[string]interface{}, with_custom_reqbody bool, req_body_custom *OAReqBodyMessageCompletion) (*OAMessage, error) {
	// send request to openai
	resp, err := c.OpenAISendMessage(content, with_format_response, format_response, with_custom_reqbody, req_body_custom)
	if err != nil {
		return nil, err
	}

	// get content first data
	data := resp.Choices[0].Message

	return &data, nil
}

// OpenAICreateImageDallE generates images based on a text prompt using either the DALL-E 2 or DALL-E 3 model.
//
// This method constructs an HTTP request to OpenAI's image generation API, validates input requirements for each model,
// and includes optional parameters for fine-tuning image quality, response format, style, size, and number of generated images.
//
// Parameters:
//
//   - req_body (*OAReqImageGeneratorDallE): A pointer to a struct containing image generation request parameters.
//
//     Fields in OAReqImageGeneratorDallE struct:
//
//   - Prompt (string): A descriptive prompt that the DALL-E model will use to generate images. This field is required.
//
//   - Model (string): Specifies the DALL-E model version, either "dall-e-2" or "dall-e-3". This is required.
//
//   - N (*int): Optional. The number of images to generate, which can range between 1 and 10. Defaults to 1 if omitted.
//
//   - Quality (*string): Optional. Available only for the DALL-E 3 model, it can be set to "standard" (default) or "hd" for high definition.
//
//   - ResponseFormat (*string): Optional. Specifies the format of the generated image response, either "url" (default) or "b64_json" for a base64-encoded image.
//
//   - Size (*string): Optional. The image size, dependent on the model:
//
//   - DALL-E 2 supports "256x256", "512x512", and "1024x1024".
//
//   - DALL-E 3 supports "1024x1024", "1792x1024", and "1024x1792".
//
//   - Style (*string): Optional. Available only for the DALL-E 3 model. Choices are "vivid" (default) or "natural" for a more realistic style.
//
//   - User (*string): Optional. A unique identifier for the end user to monitor and detect abuse, helping OpenAI with usage tracking.
//
// Returns:
//   - (*OAImageGeneratorDallEResp, error): On success, returns a pointer to an `OAImageGeneratorDallEResp` struct containing the
//     generated image details. Returns an error if any validation fails or if the request is unsuccessful.
//
// Example usage:
//
//	reqBody := &OAReqImageGeneratorDallE{
//	    Prompt: "A futuristic cityscape at sunset",
//	    Model: "dall-e-3",
//	    N: ptr(3),
//	    Quality: ptr("hd"),
//	    Style: ptr("vivid"),
//	    Size: ptr("1024x1024"),
//	    ResponseFormat: ptr("url"),
//	}
//
//	imageResp, err := apiClient.OpenAICreateImageDallE(reqBody)
//	if err != nil {
//	    log.Fatalf("Image generation failed: %v", err)
//	}
//
// Function Logic:
//  1. **Model Validation**: Ensures `Model` is either "dall-e-2" or "dall-e-3". If not, returns an error.
//  2. **N Validation**: If `N` is provided, checks if it falls between 1 and 10 (inclusive). If out of range, returns an error.
//  3. **Quality Validation**:
//     - If `Model` is "dall-e-2", `Quality` should be nil. Returns an error if a quality value is provided for this model.
//     - If `Model` is "dall-e-3", `Quality` can be "standard" or "hd". If any other value is provided, returns an error.
//  4. **Style Validation**:
//     - Ensures `Style` is only available for "dall-e-3". For "dall-e-2", returns an error if `Style` is provided.
//     - Valid values for `Style` are "vivid" or "natural"; any other value results in an error.
//  5. **ResponseFormat Validation**:
//     - If `ResponseFormat` is specified, it must be either "url" or "b64_json". Returns an error for other values.
//  6. **API Key Check**: Confirms that the API key is set; returns an error if it's empty.
//  7. **JSON Marshalling**: Serializes `req_body` to JSON format for the request body.
//  8. **Request Creation and Headers**: Sets up an HTTP POST request with necessary headers (`Content-Type` and `Authorization`).
//  9. **Response Handling**:
//     - If the HTTP response status is not 200 OK, reads and closes the response body, then returns an error.
//     - On successful response, decodes JSON data into `OAImageGeneratorDallEResp` struct and returns it.
//
// Considerations:
//   - The function relies on an HTTP client specified in the API client’s configuration (c.config.httpClient).
//   - In the event of a non-200 HTTP status, the function reads the body to allow graceful closing of the connection.
//   - OpenAI API may apply rate limiting, so ensure retry or error handling mechanisms are in place for high-frequency requests.
//
// References:
//   - OpenAI DALL E Image Generation API: https://platform.openai.com/docs/api-reference/images/create
func (c *openaiAPI) OpenAICreateImageDallE(req_body *OAReqImageGeneratorDallE) (*OAImageGeneratorDallEResp, error) {

	// ----------- input checker request
	if req_body.Model == "" || (req_body.Model != "dall-e-2" && req_body.Model != "dall-e-3") {
		return nil, errors.New("Model must be dall-e-2 or dall-e-3")
	}

	if req_body.N != nil && (*req_body.N < 1 || *req_body.N > 10) {
		return nil, errors.New("N must be between 1 and 10")
	}

	if req_body.Model != "dall-e-3" && req_body.Quality != nil {
		return nil, errors.New("Quality is only supported for dall-e-3 model")
	}

	if req_body.Quality != nil && (*req_body.Quality != "standard" && *req_body.Quality != "hd") {
		return nil, errors.New("Quality must be standard or hd")
	}

	if req_body.Model != "dall-e-3" && req_body.Style != nil {
		return nil, errors.New("Style is only supported for dall-e-3 model")
	}

	if req_body.Style != nil && (*req_body.Style != "vivid" && *req_body.Style != "natural") {
		return nil, errors.New("Style must be vivid or natural")
	}

	if req_body.ResponseFormat != nil && (*req_body.ResponseFormat != "url" && *req_body.ResponseFormat != "b64_json") {
		return nil, errors.New("ResponseFormat must be url or b64_json")
	}

	apiKey := c.apiKey
	if apiKey == "" {
		return nil, errors.New("API Key is empty")
	}

	reqBodyJson, err := json.Marshal(req_body)
	if err != nil {
		return nil, errors.New("Failed to marshal request body")
	}

	// create and send request
	req, err := http.NewRequest(http.MethodPost, OAUrlImageGenerationsDallE, bytes.NewBuffer(reqBodyJson))
	if err != nil {
		return nil, errors.New("Failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

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

	var respDataDallE OAImageGeneratorDallEResp
	if err := json.NewDecoder(resp.Body).Decode(&respDataDallE); err != nil {
		return nil, errors.New("Failed to decode response: " + err.Error())
	}

	return &respDataDallE, nil
}

// OpenAITextToSpeech converts a text input into a speech audio file using OpenAI's TTS models.
// This function validates the input parameters, prepares the request, sends it to the OpenAI API,
// and returns the audio response encoded in base64 format.
//
// Parameters:
//   - req_body (*OAReqTextToSpeech): A pointer to the OAReqTextToSpeech struct containing the TTS parameters.
//
// Returns:
//   - (*OATextToSpeechResp, error): On success, returns a pointer to an OATextToSpeechResp struct containing:
//   - FormatAudio: The file extension for the audio format (e.g., ".mp3").
//   - B64JSON: A base64-encoded string representing the audio file content.
//     On failure, returns an error.
//
// Errors:
//   - Returns an error if required fields are missing or invalid, including:
//   - Invalid Model (must be "tts-1" or "tts-1-hd").
//   - Missing Input text.
//   - Invalid Voice option (allowed values: "alloy", "echo", "fable", "onyx", "nova", "shimmer").
//   - Invalid ResponseFormat (allowed values: "mp3", "opus", "aac", "flac", "wav", "pcm").
//   - Speed out of range (0.25 to 4.0).
//   - Also returns an error if the API key is missing, or if any part of the HTTP request/response fails.
//
// Example Usage:
//
//	reqBody := OAReqTextToSpeech{
//	    Model:          "tts-1",
//	    Input:          "Hello, world!",
//	    Voice:          "alloy",
//	    ResponseFormat: "mp3",
//	}
//
//	resp, err := openAI.OpenAITextToSpeech(&reqBody)
//	if err != nil {
//	    log.Fatalf("Text-to-Speech conversion failed: %v", err)
//	}
//
//	fmt.Println("Audio Format:", resp.FormatAudio)
//	fmt.Println("Base64 Encoded Audio:", resp.B64JSON)
//
// References:
//   - TTS OpenAI: https://platform.openai.com/docs/api-reference/audio/createSpeech
func (c *openaiAPI) OpenAITextToSpeech(req_body *OAReqTextToSpeech) (*OATextToSpeechResp, error) {

	// ----------- input checker request
	if req_body.Model == "" || (req_body.Model != "tts-1" && req_body.Model != "tts-1-hd") {
		return nil, errors.New("Model must be gpt-3 or davinci")
	}

	if req_body.Input == "" {
		return nil, errors.New("Input text must be provided")
	}

	if req_body.Voice != "" && (req_body.Voice != "alloy" && req_body.Voice != "echo" && req_body.Voice != "fable" && req_body.Voice != "onyx" && req_body.Voice != "nova" && req_body.Voice != "shimer") {
		return nil, errors.New("Voice must be en or en-GB")
	}

	if req_body.ResponseFormat != "" && (req_body.ResponseFormat == "mp3" && req_body.ResponseFormat == "opus" && req_body.ResponseFormat == "aac" && req_body.ResponseFormat == "flac" && req_body.ResponseFormat == "wav" && req_body.ResponseFormat == "pcm") {
		return nil, errors.New("ResponseFormat must be mp3, opus, aac, flac, wav, or pcm")
	}

	if req_body.Speed != nil && (*req_body.Speed < 0.25 || *req_body.Speed > 4.0) {
		return nil, errors.New("Speed must be between 0.25 and 4.0")
	}

	apiKey := c.apiKey
	if apiKey == "" {
		return nil, errors.New("API Key is empty")
	}

	// create json ver for req body
	reqBodyJson, err := json.Marshal(req_body)
	if err != nil {
		return nil, errors.New("Failed to marshal request body")
	}

	// create req
	req, err := http.NewRequest(http.MethodPost, OAUrlTextToSpeech, bytes.NewBuffer(reqBodyJson))
	if err != nil {
		return nil, errors.New("Failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

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

	// decode file mp3 response to encode base64
	// because from the docs will be return file extension for audio, so for the response will be base64 encoded version of the audio we received
	var b64audio, fileExt string
	fileBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Failed to read response body3: " + err.Error())
	}

	b64audio = base64.StdEncoding.EncodeToString(fileBytes)

	if req_body.ResponseFormat == "" {
		fileExt = ".mp3"
	} else {
		fileExt = "." + req_body.ResponseFormat
	}

	result := OATextToSpeechResp{
		B64JSON:     b64audio,
		FormatAudio: fileExt,
	}

	return &result, nil
}
