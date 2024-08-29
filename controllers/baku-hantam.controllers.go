package controllers

import (
	"fmt"
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

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Topic": topicList,
	})
}

func PostBakuHantam(c *fiber.Ctx) error {

	topic := c.FormValue("topic")
	topicName := c.FormValue("topicName")

	topicData := utils.DetailBakuHantamData(topic)

	prompt := fmt.Sprintf("dibawah, saya ada data kumpulan twit dari platform X terkait topik ribut '%s', tiap data adalah satu twit dengan owner adalah pemilik twit dan quoted jika data itu ada berarti owner mengquoted twit tersebut. Coba berikan jawaban dengan dua bagian pertama analisis hal menarik kamu temukan pada kumpulan twit tersebut dan satu lagi berikan terkait rangkuman kamu terkait kumpulan twit tersebut, lakukan dengan bahasa indonesia ringan misal dengan ala jakarta lo gue dan lakukan jangan terlalu panjang, mungkin juga bisa berikan sedikit roast atau bisa sebut nama akun untuk komentari hal - hal yang menarik di twitnya dan tidak usah berikan judul bagian analisis/rangkuman pada jawabanmu. Return dalam html tag, agar dapat ditampilkan dengan rapi \ndata: ", topicName)

	prompt += fmt.Sprintf("%+v", topicData)

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
		"Data":      msg["text"],
		"TopicName": topicName,
		"Topic":     "https://bakuhantam.dev" + topic,
	})
}
