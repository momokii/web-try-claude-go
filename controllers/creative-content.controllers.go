package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"scrapper-test/models"
	"scrapper-test/utils"
	"scrapper-test/utils/openai"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type CreativeContentController struct {
	openai openai.OpenAI
}

func NewCreativeContentController(openai openai.OpenAI) *CreativeContentController {
	return &CreativeContentController{
		openai: openai,
	}
}

func (h *CreativeContentController) ViewCreativeContent(c *fiber.Ctx) error {
	return c.Render("creative-content", fiber.Map{
		"Title": "Creative Content Generator",
	})
}

func (h *CreativeContentController) GetImageAnalysis(c *fiber.Ctx) error {

	// FORM INPUT AND CHECKER
	language := c.FormValue("language", "indonesia")
	// process uploaded image to base64
	uploaded_image, err := c.FormFile("image")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "image data not found")
	}

	// --------- base variable, prompt, and format response ------------
	var contentImageAnalysisRes models.ImageAnalysisRes
	var contentRecommendationRes models.CreativeContentRecommendationRes

	prompt_image_analysis := fmt.Sprintf(`Berdasarkan gambar yang diberikan, analisis gambar tersebut dengan seksama dan berikan response dalam bahasa %s.
	
	analisis gambar tersebut pada beberapa aspek:
	1. Deskripsi gambar: berikan gambaran umum menurutmu tentang apa yang terlihat/terjadi dalam gambar.
	2. Deteksi objek: berikan informasi tentang objek-objek yang terdapat dalam gambar. Jika terdapat banyak sekali objek menurutmu, berikan maksimal 10 objek paling mencolok/menarik menurut analisis yang dilakukan.
	3. Emosi yang tersirat: berikan deskripsi tentang emosi yang ada dalam gambar tersebut. Deskripsikan secara detail dan jelas hasil analisis yang dilakukan.
	4. Elemen visual lainnya: berikan analisis tambahan tentang elemen visual lainnya yang terdapat dalam gambar tersebut contoh deskripsi ["langit biru cerah", "rumput hijau segar", "langit yang mendung berawan"]. Jika terdapat elemen visual yang menarik menurutmu, berikan deskripsi yang jelas dan detail tentang elemen visual tersebut, jika cukup banyak menurut hasil analisi, berikan maksimal 5 saja.

	Berikan response dengan kapitalisasi huruf pertama pada setiap kata agar terlihat lebih rapih untuk list Deteksi dan Elemen Visual.

	jika berdasarkan data analisis anda sebelumnya tidak ada yang menarik atau tidak informatif atau menurut anda bukan sebuah gambar yang bisa dijadikan bahan content creative, berikan response bahwa tidak ada content creative menarik yang bisa dihasilkan dari data analisis yang diberikan dengan balikan response pada kolom 'have_emotion' dengan set kolom tersebut dengan nilai 'false'. Beberapa hal yang mungkin bisa dianggap tidak menarik seperti screenshot asal, atau gambar non alam, hanya sebuah icon/logo atau apapun itu yang kamu juga lebih paham.

	Lakukan analisis gambar dengan seksama dan berikan response yang informatif dan menarik. Jika terdapat hal yang menarik atau unik dalam gambar tersebut, berikan deskripsi yang jelas dan detail tentang hal tersebut. Lakukan dengan hati - hati, detail, dan seksama.
	`, language)

	prompt_content_recommendation := fmt.Sprintf(`Berdasarkan analisis gambar yang sudah anda lakukan sebelumnya, beberapa metrik yang saya minta untuk dicari adalah terkait
	
	1. Deskripsi gambar
	2. Emosi yang tersirat dalam gambar
	3. Deteksi objek pada gambar yang saya minta jika terlalu banyak objek, berikan maksimal 10 objek yang paling mencolok
	4. Deteksi elemen visual dalam gambar yang saya minta jika terlalu banyak elemen visual, berikan maksimal 5 elemen visual yang paling mencolok

	Hasil analisis anda sebelumnya adalah data utama yang akan digunakan selanjutnya.

	Berdasarkan data analisis yang diberikan, buat beberapa output content creative yang informatif dan menarik. Content Creative bisa berupa cerita pendek, puisi, sajak, monolog, narasi singkat, atau bentuk content creative lainnya yang menurut anda sesuai dengan data analisis yang diberikan. 

	Jika anda menemukan banyak content creative, berikan maksimal 5 saja yang paling menarik menurut anda.

	Berikan tanda baca yang jelas sesuai dengan jenis konten yang anda buat.

	Pada setiap satu data output yang berikan, berikan maksimal panjang karakter yang diberikan adalah 4096 karakter dan tidak boleh lebih.

	Lakukan dengan hati - hati, detail, dan seksama.

	`)

	format_response_image_analysis := openai.OACreateResponseFormat(
		"image_analysis",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"have_emotion": map[string]interface{}{
					"type": "boolean",
				},
				"image_description": map[string]interface{}{
					"type": "string",
				},
				"emotion_detection": map[string]interface{}{
					"type": "string",
				},
				"object_detection": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"visual_element": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	)

	format_response_creative_content_maker := openai.OACreateResponseFormat(
		"creative_content_maker",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"creative_content": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"content_type": map[string]interface{}{
							"type": "string",
						},
						"content": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
	)

	// --------- main phase ------------
	// get image extension
	imageExtension := filepath.Ext(uploaded_image.Filename)
	imgExt := strings.TrimPrefix(imageExtension, ".")

	// -- convert to base64 string
	// open image
	imageFile, err := uploaded_image.Open()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "image not found")
	}
	defer func() {
		imageFile.Close()
	}()

	// convert to bytes
	imageBytes, err := io.ReadAll(imageFile)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "image not found")
	}

	// encode from bytes to base 64
	image_base64 := base64.StdEncoding.EncodeToString(imageBytes)

	// create content vision data
	messageData, err := openai.OACreateOneContentVision("image/"+imgExt, false, image_base64, prompt_image_analysis)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	// base context window model chat
	messageReq := []openai.OAMessageReq{
		{
			Role:    "user",
			Content: messageData,
		},
	}

	// send first req for image analysis
	openaiResp, err := h.openai.OpenAIGetFirstContentDataResp(&messageReq, true, &format_response_image_analysis, false, nil)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	imageAnalysisJSON := openaiResp.Content // get json response data

	// decode model json response to struct
	if err := json.NewDecoder(strings.NewReader(imageAnalysisJSON)).Decode(&contentImageAnalysisRes); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	// SECOND REQUEST TO MODEL AFTER GET THE IMAGE ANALYSIS
	// do the 2nd JUST IF the image analysis result say the image "have_emotion" is true
	if contentImageAnalysisRes.HaveEmotion {
		// make chaining request data, add image anaylsis result and prompt content recommendation for next request
		messageReq = append(messageReq, openai.OAMessageReq{
			Role:    "assistant",
			Content: imageAnalysisJSON,
		})

		messageReq = append(messageReq, openai.OAMessageReq{
			Role:    "user",
			Content: prompt_content_recommendation,
		})

		// send 2nd req
		openaiResp, err = h.openai.OpenAIGetFirstContentDataResp(&messageReq, true, &format_response_creative_content_maker, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}

		creative_content_recommendation := openaiResp.Content

		if err = json.NewDecoder(strings.NewReader(creative_content_recommendation)).Decode(&contentRecommendationRes); err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "list analysis images", fiber.Map{
		"analysis":               contentImageAnalysisRes,
		"content_recommendation": contentRecommendationRes,
	})
}

func (h *CreativeContentController) CreateImageDallE(c *fiber.Ctx) error {

	userInput := new(models.CreateImageGenerator) // struct for user input
	if err := c.BodyParser(&userInput); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	size := "1792x1024"
	response := "b64_json"
	imageReqBody := openai.OAReqImageGeneratorDallE{
		Prompt:         userInput.Prompt,
		Model:          "dall-e-3",
		Size:           &size,
		ResponseFormat: &response,
	}
	imageData, err := h.openai.OpenAICreateImageDallE(&imageReqBody)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "image generator success", fiber.Map{
		"image_data": imageData,
	})
}

func (h *CreativeContentController) CreateTTS(c *fiber.Ctx) error {

	userInput := new(models.CreateTTS)
	if err := c.BodyParser(&userInput); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	var prompt string

	prompt += userInput.Prompt
	ttsReqBody := openai.OAReqTextToSpeech{
		Model:          "tts-1",
		Input:          prompt,
		Voice:          "alloy",
		ResponseFormat: "mp3",
	}

	ttsData, err := h.openai.OpenAITextToSpeech(&ttsReqBody)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "text to speech success", fiber.Map{
		"audio_format": ttsData.FormatAudio,
		"b64_json":     ttsData.B64JSON,
	})
}
