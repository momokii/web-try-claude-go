package controllers

import (
	"encoding/json"
	"fmt"
	"scrapper-test/models"
	"scrapper-test/utils"
	"scrapper-test/utils/claude"
	"scrapper-test/utils/openai"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type StoriesController struct {
	claude claude.ClaudeAPI
	openai openai.OpenAI
}

func NewStoriesController(claude claude.ClaudeAPI, openai openai.OpenAI) *StoriesController {
	return &StoriesController{
		claude: claude,
		openai: openai,
	}
}

func (h *StoriesController) ViewStories(c *fiber.Ctx) error {
	return c.Render("stories", fiber.Map{
		"Title": "Create Your Own Stories",
	})
}

func (h *StoriesController) CreateStoriesTitle(c *fiber.Ctx) error {

	var parsedResponse models.StoriesCreateTitleFormat
	var jsonResp string

	// get model query to determine which model to use
	type_llm := c.Query("model")

	inputUser := new(models.StoriesCreateInput)
	if err := c.BodyParser(inputUser); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	prompt := fmt.Sprintf(`Berdasarkan tema ['%s'], hasilkan 4 judul cerita pendek yang menarik dan dalam bahasa ['%s'] juga cerita terkait cerita yang ada di ['%s']. Berikan deskripsi sederhana dengan 1-2 kalimat.
	
	Berikan format judul dengan "NAMA JUDUL" tanpa "a. NAMA JUDUL" atau "1. NAMA JUDUL"

	`, inputUser.Theme, inputUser.Language, inputUser.Language)

	if type_llm == "claude" {
		// if using claude add json response format in prompt
		prompt += `
			Berikan jawaban dalam struktur response API JSON penuh dan berikan jawaban hanya struktur JSON saja dengan struktur

			{"titles": [{"title", "description"}]}
		`

		prompt_input := []claude.ClaudeMessageReq{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		claudeRes, err := h.claude.ClaudeGetFirstContentDataResp(&prompt_input, 10*512, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}

		jsonResp = claudeRes.Text

	} else {
		prompt_gpt := []openai.OAMessageReq{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		// response format for openai
		format_response := openai.OACreateResponseFormat(
			"titles_choices",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"titles": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"title": map[string]string{
									"type": "string",
								},
								"description": map[string]string{
									"type": "string",
								},
							},
						},
					},
				},
			},
		)

		openairesp, err := h.openai.OpenAIGetFirstContentDataResp(&prompt_gpt, true, &format_response, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "openai: "+err.Error())
		}

		jsonResp = openairesp.Content
	}

	// decode response from openai
	if err := json.NewDecoder(strings.NewReader(jsonResp)).Decode(&parsedResponse); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "create stories title", fiber.Map{
		"titles": parsedResponse.Titles,
	})
}

func (h *StoriesController) CreateFirstStoriesPart(c *fiber.Ctx) error {

	var parsedResponse models.StoriesCreateParagraph
	var jsonResp string

	type_llm := c.Query("model")

	inputUser := new(models.StoriesCreateFirstPartInput)
	if err := c.BodyParser(inputUser); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	prompt := fmt.Sprintf(`Berdasarkan judul yang dipilih ['%s'] dengan tema ['%s'] dan deskripsi ['%s'], hasilkan awal cerita pendek yang menarik dalam bahasa ['%s'] berikan dalam 3-4 kalimat diakhiri dengan keadaan yang membutuhkan keputusan.

	Return pada paragraf hanya berisi paragraf baru saja tanpa pilihan keputusan baru yang akan digunakan dan juga tanpa seperti '\n' dan sejenisnya. Jika diperlukan berikan input tersebut dalam tag HTML

	Kemudian berikan 4 pilihan keputusan yang bisa diambil oleh karakter utama untuk dapat melanjutkan cerita.

	Berikan format keputusan dalam array dengan ["keputusan 1", "keputusan -n"] tanpa "a. KEPUTUSAN" atau "1. KEPUTUSAN"

	`, inputUser.Title, inputUser.Theme, inputUser.Description, inputUser.Language)

	if type_llm == "claude" {
		prompt += `
		berikan format jawaban hanya struktur JSON saja dengan struktur diberikan

		{ "paragraph", "choices" : ["choice"]}
		`

		prompt_input := []claude.ClaudeMessageReq{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		claudeRes, err := h.claude.ClaudeGetFirstContentDataResp(&prompt_input, 512*10, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}

		jsonResp = claudeRes.Text
	} else {

		prompt_input := []openai.OAMessageReq{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		response_format := openai.OACreateResponseFormat(
			"paragraph_choices",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"paragraph": map[string]string{
						"type": "string",
					},
					"choices": map[string]interface{}{
						"type": "array",
						"items": map[string]string{
							"type": "string",
						},
					},
				},
			},
		)

		gptRes, err := h.openai.OpenAIGetFirstContentDataResp(&prompt_input, true, &response_format, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}

		jsonResp = gptRes.Content
	}

	if err := json.NewDecoder(strings.NewReader(jsonResp)).Decode(&parsedResponse); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "create stories first part", fiber.Map{
		"paragraph": parsedResponse.Paragraph,
		"choices":   parsedResponse.Choices,
	})
}

func (h *StoriesController) CreateStoriesParagraph(c *fiber.Ctx) error {

	var parsedResponse models.StoriesCreateParagraph
	var prompt, jsonResp string

	type_llm := c.Query("model")

	data := c.Params("data")
	if data != "next" {
		data = "end"
	}

	inputUser := new(models.StoriesCreateParagraphContinueInput)
	if err := c.BodyParser(inputUser); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if data == "next" {
		prompt = fmt.Sprintf(`Berdasarkan cerita pendek bersambung yang sedang dibuat dengan data sebelumnya yang sudah didapat. Lanjutkan cerita berikut dengan mempertimbangkan pilihan yang diambil. 

		Judul: '%s'
		Deskripsi: '%s'
		Tema:'%s'
		Bahasa penulisan: '%s'
		Paragraph sampai saat ini:
		'%s'

		Pilihan yang diambil:'%s'

		Buatlah paragraf lanjutan (3-4 kalimat) yang menggambarkan konsekuensi dari pilihan tersebut diakhiri dengan situasi baru yang membutuhkan keputusan.
		
		Kemudian berikan 4 pilihan keputusan baru yang dapat diambil oleh karakter utama.

		Berikan format keputusan dalam array dengan ["keputusan 1", "keputusan -n"] tanpa "a. KEPUTUSAN" atau "1. KEPUTUSAN"

		Return pada paragraf hanya berisi paragraf baru saja tanpa pilihan keputusan baru yang akan digunakan dan juga tanpa seperti '\n' dan sejenisnya. Jika diperlukan berikan input tersebut dalam tag HTML

		`, inputUser.Title, inputUser.Description, inputUser.Theme, inputUser.Language, inputUser.Paragraph, inputUser.Choice)
	} else {
		prompt = fmt.Sprintf(`Berdasarkan cerita pendek bersambung yang sedang dibuat dengan data sebelumnya yang sudah didapat. Ini merupakan bagian akhir cerita.  Berdasarkan seluruh cerita dan pilihan terakhir yang diambil, buatlah paragraf penutup yang memberikan kesimpulan yang memuaskan.

		Judul: '%s'
		Deskripsi: '%s'
		Tema:'%s'
		Bahasa penulisan: '%s'
		Paragraph sampai saat ini:
		'%s'

		Pilihan yang diambil:'%s'

		Buatlah paragraf akhir(3-4 kalimat per paragraf) menggambarkan konsekuensi dari pilihan yang dipilih. Jika merasa hasil kurang baik untuk penutup yang memuaskan bisa tambahkan lebih dari satu (1) paragraf.

		Jika lebih dari 1 paragraf, jeda paragraf tandai dengan <br> tag

		Return pada paragraf hanya berisi paragraf baru saja tanpa pilihan keputusan baru yang akan digunakan dan juga tanpa seperti '\n' dan sejenisnya. Jika diperlukan berikan input tersebut dalam tag HTML

		`, inputUser.Title, inputUser.Description, inputUser.Theme, inputUser.Language, inputUser.Paragraph, inputUser.Choice)
	}

	if type_llm == "claude" {
		if data == "next" {
			prompt += `
			Berikan format jawaban hanya struktur JSON saja dengan struktur seperti di bawah dan pada paragraph hanya berisi ceritan penutup.

			Berikan format jawaban hanya struktur JSON saja dengan struktur seperti di bawah dan pada paragraph hanya berisi ceritan lanjutan barunya tanpa inputan paragraph yang diberikan di atas.

			{"paragraph", "choices" : ["choice"]}
			`

		} else {
			prompt += `
			tetap berikan jawaban "choices" namun berikan dengan nilai list kosong []

			{"paragraph", "choices" : []}
			`
		}

		prompt_input := []claude.ClaudeMessageReq{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		claudeRes, err := h.claude.ClaudeGetFirstContentDataResp(&prompt_input, 512*10, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}

		jsonResp = claudeRes.Text
	} else {

		prompt_input := []openai.OAMessageReq{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		response_format := openai.OACreateResponseFormat(
			"paragraph_choices",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"paragraph": map[string]string{
						"type": "string",
					},
					"choices": map[string]interface{}{
						"type": "array",
						"items": map[string]string{
							"type": "string",
						},
					},
				},
			},
		)

		gptRes, err := h.openai.OpenAIGetFirstContentDataResp(&prompt_input, true, &response_format, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}

		jsonResp = gptRes.Content
	}

	if err := json.NewDecoder(strings.NewReader(jsonResp)).Decode(&parsedResponse); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "create stories paragraph", fiber.Map{
		"paragraph": parsedResponse.Paragraph,
		"choices":   parsedResponse.Choices,
	})
}
