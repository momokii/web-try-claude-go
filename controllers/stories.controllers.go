package controllers

import (
	"encoding/json"
	"fmt"
	"scrapper-test/models"
	"scrapper-test/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ViewStories(c *fiber.Ctx) error {
	return c.Render("stories", fiber.Map{
		"Title": "Create Your Own Stories",
	})
}

func CreateStoriesTitle(c *fiber.Ctx) error {

	var parsedResponse models.StoriesCreateTitleFormat

	inputUser := new(models.StoriesCreateInput)
	if err := c.BodyParser(inputUser); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	prompt := fmt.Sprintf(`Berdasarkan tema ['%s'], hasilkan 4 judul cerita pendek yang menarik dan dalam bahasa ['%s'] juga cerita terkait cerita yang ada di ['%s'].

	Berikan jawaban dalam struktur response API JSON penuh dan berikan jawaban hanya struktur JSON saja dengan struktur

	{"titles": [{"title", "description"}]}`, inputUser.Theme, inputUser.Language, inputUser.Language)

	prompt_input := []models.ClaudeMessageReq{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	claudeRes, err := utils.ClaudeGetContentDataResp(prompt_input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	content, ok := claudeRes["text"].(string)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "claude response not found")
	}

	if err = json.NewDecoder(strings.NewReader(content)).Decode(&parsedResponse); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "create stories title", fiber.Map{
		"titles": parsedResponse.Titles,
	})
}

func CreateFirstStoriesPart(c *fiber.Ctx) error {

	var parsedResponse models.StoriesCreateParagraph

	inputUser := new(models.StoriesCreateFirstPartInput)
	if err := c.BodyParser(inputUser); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	prompt := fmt.Sprintf(`Berdasarkan judul yang dipilih ['%s'] dengan tema ['%s'] dan deskripsi ['%s'], hasilkan awal cerita pendek yang menarik dalam bahasa ['%s'] berikan dalam 3-4 kalimat diakhiri dengan keadaan yang membutuhkan keputusan.

	Setelah paragraf berikan 4 pilihan keputusan yang bisa diambil oleh karakter utama

	berikan format jawaban hanya struktur JSON saja dengan struktur diberikan

	{ "paragraph", "choices" : ["choice"]}`, inputUser.Title, inputUser.Theme, inputUser.Description, inputUser.Language)

	prompt_input := []models.ClaudeMessageReq{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	claudeRes, err := utils.ClaudeGetContentDataResp(prompt_input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	content, ok := claudeRes["text"].(string)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "claude response not found")
	}

	if err = json.NewDecoder(strings.NewReader(content)).Decode(&parsedResponse); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "create stories first part", fiber.Map{
		"paragraph": parsedResponse.Paragraph,
		"choices":   parsedResponse.Choices,
	})
}

func CreateStoriesParagraph(c *fiber.Ctx) error {

	var parsedResponse models.StoriesCreateParagraph
	var prompt string
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

		Buatlah paragraf lanjutan (3-4 kalimat) yang menggambarkan konsekuensi dari pilihan tersebut diakhiri dengan situasi baru yang membutuhkan keputusan. Setelah paragraf berikan 4 pilihan keputusan baru.

		Berikan format jawaban hanya struktur JSON saja dengan struktur seperti di bawah dan pada paragraph hanya berisi ceritan lanjutan barunya tanpa inputan paragraph yang diberikan di atas.

		{"paragraph", "choices" : ["choice"]}
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

		Berikan format jawaban hanya struktur JSON saja dengan struktur seperti di bawah dan pada paragraph hanya berisi ceritan penutup.

		Jika lebih dari 1 paragraf, jeda paragraf tandai dengan 2 <br> tag

		tetap berikan jawaban "choices" namun berikan dengan nilai list kosong []

		{"paragraph", "choices" : []}
		`, inputUser.Title, inputUser.Description, inputUser.Theme, inputUser.Language, inputUser.Paragraph, inputUser.Choice)
	}

	prompt_input := []models.ClaudeMessageReq{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	claudeRes, err := utils.ClaudeGetContentDataResp(prompt_input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	content, ok := claudeRes["text"].(string)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "claude response not found")
	}

	if err = json.NewDecoder(strings.NewReader(content)).Decode(&parsedResponse); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "create stories paragraph", fiber.Map{
		"paragraph": parsedResponse.Paragraph,
		"choices":   parsedResponse.Choices,
	})
}
