package controllers

import (
	"fmt"
	"scrapper-test/models"
	"scrapper-test/utils"

	"github.com/gofiber/fiber/v2"
)

func ViewBakuHantam(c *fiber.Ctx) error {
	return c.Render("baku-hantam", fiber.Map{
		"Title": "Bakun Hantam X",
	})
}

func GetBakuHantamTopic(c *fiber.Ctx) error {
	topicList := utils.GetBakuHantamTopic()
	return utils.ResponseWithData(c, fiber.StatusOK, "List of Bakuhantam Topic", fiber.Map{
		"topic": topicList,
	})
}

func PostBakuHantam(c *fiber.Ctx) error {

	topic := c.FormValue("topic")
	topicName := c.FormValue("topicName")

	topicData := utils.DetailBakuHantamData(topic)

	prompt := fmt.Sprintf("dibawah, saya ada data kumpulan twit dari platform X terkait topik ribut '%s', tiap data adalah satu twit dengan owner adalah pemilik twit dan quoted jika data itu ada berarti owner mengquoted twit tersebut. Coba berikan jawaban dengan dua bagian pertama analisis hal menarik kamu temukan pada kumpulan twit tersebut dan satu lagi berikan terkait rangkuman kamu terkait kumpulan twit tersebut, lakukan dengan bahasa indonesia ringan misal dengan ala jakarta lo gue dan lakukan jangan terlalu panjang, mungkin juga bisa berikan sedikit roast atau bisa sebut nama akun untuk komentari hal - hal yang menarik di twitnya dan tidak usah berikan judul bagian analisis/rangkuman pada jawabanmu. Return dalam html tag, agar dapat ditampilkan dengan rapi \ndata: ", topicName)

	prompt += fmt.Sprintf("%+v", topicData)

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
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "claude response not found")
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "Bakuhantam Response", fiber.Map{
		"content":    content,
		"topic_name": topicName,
		"topic":      "https://bakuhantam.dev" + topic,
	})
}
