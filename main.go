package main

import (
	"net/http"
	"os"
	"scrapper-test/controllers"
	"scrapper-test/database"
	"scrapper-test/middlewares"
	"scrapper-test/models"
	"scrapper-test/repository/session"
	"scrapper-test/repository/user"
	"scrapper-test/utils/claude"
	"scrapper-test/utils/openai"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	engine := html.New("./public", ".html")
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	claude, err := claude.New(
		os.Getenv("CLAUDE_API_KEY"),
		claude.WithHTTPClient(httpClient),
		claude.WithBaseUrl(os.Getenv("CLAUDE_BASE_URL")),
		claude.WithModel(os.Getenv("CLAUDE_MODEL")),
		claude.WithAnthropicVersion(os.Getenv("CLAUDE_ANTHROPIC_VERSION")),
	)
	if err != nil {
		panic(err)
	}

	openai, err := openai.New(
		os.Getenv("OA_APIKEY"),
		os.Getenv("OA_ORGANIZATIONID"),
		os.Getenv("OA_PROJECTID"),
		openai.WithHTTPClient(httpClient),
		openai.WithModel("gpt-4o"),
		openai.WithBaseUrl("https://api.openai.com/v1/chat/completions"),
	)
	if err != nil {
		panic(err)
	}

	// db and session storage init
	database.InitDB()
	middlewares.InitSession()

	// repo init
	userRepo := user.NewUserRepo()
	sessionRepo := session.NewSessionRepo()

	// controller
	mediumController := controllers.NewMediumController(claude, openai)
	bakuHantamController := controllers.NewBakuHantamController(claude, openai)
	storiesController := controllers.NewStoriesController(claude, openai)
	creativecontentController := controllers.NewCreativeContentController(openai)
	authHandler := controllers.NewAuthHandler(*userRepo, *sessionRepo)

	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).Render("errorPage", fiber.Map{
				"Title": "Error",
				"Error": err.Error(),
				"Code":  code,
			})
		},
	})
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(helmet.New())
	app.Use(recover.New())
	// not using rate limiter for now
	app.Use(limiter.New(limiter.Config{
		Max:               30,
		Expiration:        1 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{}, // sliding window rate limiter,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).Render("errorPage", fiber.Map{
				"Title":   "Error",
				"message": "kebanyakan riques bre, balik lagi ntar yak",
				"Code":    fiber.StatusTooManyRequests,
			})
		},
	}))
	app.Static("/public", "./public")

	// route
	app.Get("/", middlewares.IsAuth, func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.UserSession)

		return c.Render("index", fiber.Map{
			"Title": "Hello, LLM!",
			"User":  user,
		})
	})

	// auth sso
	app.Get("/auth/sso", middlewares.IsNotAuth, authHandler.SSOAuthLogin)
	app.Post("/api/logout", middlewares.IsAuth, authHandler.Logout)

	app.Get("/medium", middlewares.IsAuth, mediumController.ViewMedium)
	app.Post("/api/medium", middlewares.IsAuth, mediumController.PostMedium)

	app.Get("/baku-hantam", middlewares.IsAuth, bakuHantamController.ViewBakuHantam)
	app.Post("/api/baku-hantam", middlewares.IsAuth, bakuHantamController.PostBakuHantam)
	app.Get("/api/baku-hantam/topics", middlewares.IsAuth, bakuHantamController.GetBakuHantamTopic)

	app.Get("/stories", middlewares.IsAuth, storiesController.ViewStories)
	app.Post("/api/stories/titles", middlewares.IsAuth, storiesController.CreateStoriesTitle)
	app.Post("/api/stories/paragraphs", middlewares.IsAuth, storiesController.CreateFirstStoriesPart)
	app.Post("/api/stories/paragraphs/:data", middlewares.IsAuth, storiesController.CreateStoriesParagraph)

	app.Get("/creative-content", middlewares.IsAuth, creativecontentController.ViewCreativeContent)
	app.Post("/api/creative-content/images/analysis", middlewares.IsAuth, creativecontentController.GetImageAnalysis)
	app.Post("/api/creative-content/images/generations", middlewares.IsAuth, creativecontentController.CreateImageDallE)
	app.Post("/api/creative-content/audio/speech", middlewares.IsAuth, creativecontentController.CreateTTS)

	app.Listen(":3002")
}
