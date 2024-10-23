package controllers

import (
	"scrapper-test/utils"
	"scrapper-test/utils/claude"
	"scrapper-test/utils/openai"

	"github.com/gofiber/fiber/v2"
)

type mediumController struct {
	claude claude.ClaudeAPI
	openai openai.OpenAI
}

func NewMediumController(claude claude.ClaudeAPI, openai openai.OpenAI) *mediumController {
	return &mediumController{
		claude: claude,
		openai: openai,
	}
}

func (h *mediumController) ViewMedium(c *fiber.Ctx) error {
	return c.Render("medium", fiber.Map{
		"Title": "Medium Roasting",
	})
}

func (h *mediumController) PostMedium(c *fiber.Ctx) error {

	var content string

	username := c.FormValue("username")
	llm_type := c.FormValue("model")

	mediumData := utils.MediumProfileScrapper(username)

	prompt := `
	Berikan roasting playful untuk konten Medium user berikut dengan kriteria:
	- Gaya bahasa: Santai/gaul Jakarta (lo-gue)
	- Tone: Playful tapi savage 
	- Panjang: 2-3 paragraf max
	- Focus roasting pada:
	* Topic/niche yang dipilih author
	* Writing style & clickbait level
	* Konsistensi posting
	* Engagement & kualitas konten
	* Fun fact atau pattern menarik

	Note: Data post diambil max 10 tulisan terakhir per user. Tidak perlu mention jumlah post jika tepat 10.

	Data Medium:
	` + mediumData.PromptData

	if llm_type == "claude" {
		prompt_input := []claude.ClaudeMessageReq{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		claudeResp, err := h.claude.ClaudeGetFirstContentDataResp(prompt_input, 256*10)
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

		openaiResp, err := h.openai.OpenAIGetFirstContentDataResp(prompt_input, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}

		content = openaiResp.Content
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "medium data roasting", fiber.Map{
		"profile": mediumData.MediumProfileUser,
		"content": content,
	})
}
