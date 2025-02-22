package controllers

import (
	"log"
	"scrapper-test/database"
	"scrapper-test/utils"
	"scrapper-test/utils/claude"
	"scrapper-test/utils/openai"

	sso_models "github.com/momokii/go-sso-web/pkg/models"
	sso_user "github.com/momokii/go-sso-web/pkg/repository/user"
	sso_utils "github.com/momokii/go-sso-web/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type mediumController struct {
	claude   claude.ClaudeAPI
	openai   openai.OpenAI
	userRepo sso_user.UserRepo
}

func NewMediumController(claude claude.ClaudeAPI, openai openai.OpenAI, userRepo sso_user.UserRepo) *mediumController {
	return &mediumController{
		claude:   claude,
		openai:   openai,
		userRepo: userRepo,
	}
}

func (h *mediumController) ViewMedium(c *fiber.Ctx) error {
	return c.Render("medium", fiber.Map{
		"Title": "Medium Roasting",
	})
}

func (h *mediumController) PostMedium(c *fiber.Ctx) error {

	// check if user is exist
	user_session := c.Locals("user").(sso_models.UserSession)

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	defer func() {
		database.CommitOrRollback(tx, c, err)
	}()

	user, err := h.userRepo.FindByID(tx, user_session.Id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	if user.Id == 0 {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "user not found")
	}

	// check if user have enough credit token
	if user.CreditToken < utils.FEATURE_MEDIUM_COST {
		log.Println(user)
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Not enough credit token to use this feature")
	}

	// start process and using the FEATURE

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

		openaiResp, err := h.openai.OpenAIGetFirstContentDataResp(&prompt_input, false, nil, false, nil)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}

		content = openaiResp.Content
	}

	// feature success executed, reduce user credit token
	if err := sso_utils.UpdateUserCredit(tx, h.userRepo, user, utils.FEATURE_MEDIUM_COST); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseWithData(c, fiber.StatusOK, "medium data roasting", fiber.Map{
		"profile": mediumData.MediumProfileUser,
		"content": content,
	})
}
