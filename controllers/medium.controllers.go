package controllers

import (
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

	claudeResp, err := utils.SendOneMessage(prompt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	claudeRespMsg, ok := claudeResp.(map[string]interface{})["content"].([]interface{})
	if !ok || len(claudeRespMsg) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "claude response not found",
		})
	}

	msg, ok := claudeRespMsg[0].(map[string]interface{})
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "error mashe",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Profile": mediumData.MediumProfileUser,
		"Data":    msg["text"],
	})
}
