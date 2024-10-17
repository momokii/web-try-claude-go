package controllers

import (
	"scrapper-test/models"
	"scrapper-test/utils"

	"github.com/gofiber/fiber/v2"
)

func ViewMedium(c *fiber.Ctx) error {
	return c.Render("medium", fiber.Map{
		"Title": "Medium Roasting",
	})
}

func PostMedium(c *fiber.Ctx) error {

	username := c.FormValue("username")

	mediumData := utils.MediumProfileScrapper(username)

	prompt := "roast akun medium dengan data di bawah ini, lakukan dengan bahasa indonesia ala jakarta dengan lo gue dan lakukan jangan terlalu panjang, berikan jawaban hanya hasil roast dan konteks untuk data post medium memang hanya diambil maksimal 10 jika ada lebih dari data user jadi jika data post adalah 10 tidak usah bahas jumlahnya. \ndata: " + mediumData.PromptData

	prompt_input := []models.ClaudeMessageReq{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	claudeResp, err := utils.ClaudeGetContentDataResp(prompt_input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	content, ok := claudeResp["text"].(string)
	if !ok || len(content) == 0 {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "claude response not found")
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "medium data roasting", fiber.Map{
		"profile": mediumData.MediumProfileUser,
		"content": content,
	})
}
