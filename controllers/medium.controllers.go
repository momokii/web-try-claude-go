package controllers

import (
	"scrapper-test/utils"
	"scrapper-test/utils/claude"

	"github.com/gofiber/fiber/v2"
)

type mediumController struct {
	claude claude.ClaudeAPI
}

func NewMediumController(claude claude.ClaudeAPI) *mediumController {
	return &mediumController{
		claude: claude,
	}
}

func (h *mediumController) ViewMedium(c *fiber.Ctx) error {
	return c.Render("medium", fiber.Map{
		"Title": "Medium Roasting",
	})
}

func (h *mediumController) PostMedium(c *fiber.Ctx) error {

	username := c.FormValue("username")

	mediumData := utils.MediumProfileScrapper(username)

	prompt := "roast akun medium dengan data di bawah ini, lakukan dengan bahasa indonesia ala jakarta dengan lo gue dan lakukan jangan terlalu panjang, berikan jawaban hanya hasil roast dan konteks untuk data post medium memang hanya diambil maksimal 10 jika ada lebih dari data user jadi jika data post adalah 10 tidak usah bahas jumlahnya. \ndata: " + mediumData.PromptData

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

	content := claudeResp.Text

	return utils.ResponseWithData(c, fiber.StatusOK, "medium data roasting", fiber.Map{
		"profile": mediumData.MediumProfileUser,
		"content": content,
	})
}
