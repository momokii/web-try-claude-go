package main

import (
	"net/http"
	"os"
	"scrapper-test/controllers"
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

	// controller
	mediumController := controllers.NewMediumController(claude, openai)
	bakuHantamController := controllers.NewBakuHantamController(claude, openai)
	storiesController := controllers.NewStoriesController(claude, openai)

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
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Hello, LLM!",
		})
	})

	app.Get("/medium", mediumController.ViewMedium)
	app.Post("/api/medium", mediumController.PostMedium)

	app.Get("/baku-hantam", bakuHantamController.ViewBakuHantam)
	app.Post("/api/baku-hantam", bakuHantamController.PostBakuHantam)
	app.Get("/api/baku-hantam/topics", bakuHantamController.GetBakuHantamTopic)

	app.Get("/stories", storiesController.ViewStories)
	app.Post("/api/stories/titles", storiesController.CreateStoriesTitle)
	app.Post("/api/stories/paragraphs", storiesController.CreateFirstStoriesPart)
	app.Post("/api/stories/paragraphs/:data", storiesController.CreateStoriesParagraph)

	app.Listen(":3000")
}
