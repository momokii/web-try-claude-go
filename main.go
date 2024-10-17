package main

import (
	"scrapper-test/controllers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	engine := html.New("./public", ".html")

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
	// app.Use(limiter.New(limiter.Config{
	// 	Max:               15,
	// 	Expiration:        1 * time.Minute,
	// 	LimiterMiddleware: limiter.SlidingWindow{}, // sliding window rate limiter,
	// 	LimitReached: func(c *fiber.Ctx) error {
	// 		return c.Status(fiber.StatusTooManyRequests).Render("errorPage", fiber.Map{
	// 			"Title":   "Error",
	// 			"message": "kebanyakan riques bre, balik lagi ntar yak",
	// 			"Code":    fiber.StatusTooManyRequests,
	// 		})
	// 	},
	// }))
	app.Static("/public", "./public")

	// route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Hello, Sonnet!",
		})
	})

	app.Get("/medium", controllers.ViewMedium)
	app.Post("/api/medium", controllers.PostMedium)

	app.Get("/baku-hantam", controllers.ViewBakuHantam)
	app.Post("/api/baku-hantam", controllers.PostBakuHantam)
	app.Get("/api/baku-hantam/topics", controllers.GetBakuHantamTopic)

	app.Get("/stories", controllers.ViewStories)
	app.Post("/api/stories/titles", controllers.CreateStoriesTitle)
	app.Post("/api/stories/paragraphs", controllers.CreateFirstStoriesPart)
	app.Post("/api/stories/paragraphs/:data", controllers.CreateStoriesParagraph)

	app.Listen(":3000")

}
