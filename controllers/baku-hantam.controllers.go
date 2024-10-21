package controllers

import (
	"fmt"
	"scrapper-test/utils"
	"scrapper-test/utils/claude"

	"github.com/gofiber/fiber/v2"
)

type BakuHantamController struct {
	claude claude.ClaudeAPI
}

func NewBakuHantamController(claude claude.ClaudeAPI) *BakuHantamController {
	return &BakuHantamController{
		claude: claude,
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

	topic := c.FormValue("topic")
	topicName := c.FormValue("topicName")

	topicData := utils.DetailBakuHantamData(topic)

	prompt := fmt.Sprintf("dibawah, saya ada data kumpulan twit dari platform X terkait topik ribut '%s', tiap data adalah satu twit dengan owner adalah pemilik twit dan quoted jika data itu ada berarti owner mengquoted twit tersebut. Coba berikan jawaban dengan dua bagian pertama analisis hal menarik kamu temukan pada kumpulan twit tersebut dan satu lagi berikan terkait rangkuman kamu terkait kumpulan twit tersebut, lakukan dengan bahasa indonesia ringan misal dengan ala jakarta lo gue dan lakukan jangan terlalu panjang, mungkin juga bisa berikan sedikit roast atau bisa sebut nama akun untuk komentari hal - hal yang menarik di twitnya dan tidak usah berikan judul bagian analisis/rangkuman pada jawabanmu. Return dalam html tag, agar dapat ditampilkan dengan rapi \ndata: ", topicName)

	prompt += fmt.Sprintf("%+v", topicData)

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

	return utils.ResponseWithData(c, fiber.StatusOK, "Bakuhantam Response", fiber.Map{
		"content":    content,
		"topic_name": topicName,
		"topic":      "https://bakuhantam.dev" + topic,
	})
}
