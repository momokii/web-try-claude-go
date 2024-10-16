package utils

import "github.com/gofiber/fiber/v2"

func ErrorResponse(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": message,
	})
}

func ResponseMessage(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(fiber.Map{
		"error":   false,
		"message": message,
	})
}

func ResponseWithData(c *fiber.Ctx, code int, message string, data interface{}) error {
	return c.Status(code).JSON(fiber.Map{
		"error":   false,
		"message": message,
		"data":    data,
	})
}
