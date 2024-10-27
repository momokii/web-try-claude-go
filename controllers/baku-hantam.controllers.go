package controllers

import (
	"fmt"
	"scrapper-test/utils"
	"scrapper-test/utils/claude"
	"scrapper-test/utils/openai"

	"github.com/gofiber/fiber/v2"
)

type BakuHantamController struct {
	claude claude.ClaudeAPI
	openai openai.OpenAI
}

func NewBakuHantamController(claude claude.ClaudeAPI, openai openai.OpenAI) *BakuHantamController {
	return &BakuHantamController{
		claude: claude,
		openai: openai,
	}
}

func (h *BakuHantamController) ViewBakuHantam(c *fiber.Ctx) error {
	return c.Render("baku-hantam", fiber.Map{
		"Title": "Bakun Hantam X",
	})
}

func (h *BakuHantamController) GetBakuHantamTopic(c *fiber.Ctx) error {
	topicList := utils.GetBakuHantamTopic()
	return utils.ResponseWithData(c, fiber.StatusOK, "List of Bakuhantam Topic", fiber.Map{
		"topic": topicList,
	})
}

func (h *BakuHantamController) PostBakuHantam(c *fiber.Ctx) error {

	var content string

	topic := c.FormValue("topic")
	topicName := c.FormValue("topicName")
	type_llm := c.FormValue("model")

	topicData := utils.DetailBakuHantamData(topic)

	prompt := fmt.Sprintf(`
	Analisis kumpulan tweet dari X tentang topik '%s'. Data berisi tweet individual dengan informasi owner (pemilik tweet) dan quoted (jika tweet tersebut mengutip tweet lain).

	Berikan 2 bagian analisis dengan gaya bahasa santai/gaul (ala Jakarta):

	1. Highlight & Analisis (maks 3-4 paragraf):
	- Temuan menarik/unik dari tweet-tweet tersebut
	- Pattern atau tren yang terlihat
	- Interaksi antar user yang eye-catching
	- Feel free buat roasting secara playful ke tweet/user tertentu yang mencolok
	- Sebutkan username spesifik kalau relevan

	2. TL;DR / Ringkasan (1-2 paragraf):
	- Intisari dari drama/discourse yang terjadi
	- Tone & sentiment dominan dari percakapan
	- Quick take kamu tentang topik ini overall

	Format output dalam HTML tags untuk readability. Hindari penggunaan header/judul section yaitu tidak perlu ditulis (Highlight/ Ringkasan) sebagai pembuka paragraf. Ketika terdapat beda paragraf gunakan <br> untuk line break.

	Data tweet:

	'%v'
	`, topicName, topicData)

	if type_llm == "claude" {
		prompt_input := []claude.ClaudeMessageReq{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		claudeResp, err := h.claude.ClaudeGetFirstContentDataResp(&prompt_input, 256*10, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}

		content = claudeResp.Text

	} else {
		prompt_input := []openai.OAMessageReq{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		gptResp, err := h.openai.OpenAIGetFirstContentDataResp(&prompt_input, false, nil, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}

		content = gptResp.Content
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "Bakuhantam Response", fiber.Map{
		"content":    content,
		"topic_name": topicName,
		"topic":      "https://bakuhantam.dev" + topic,
	})
}
