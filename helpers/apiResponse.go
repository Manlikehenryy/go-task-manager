package helpers

import (
    "github.com/gofiber/fiber/v2"
)

func SendJSON(c *fiber.Ctx, statusCode int, data interface{}) error {
    return c.Status(statusCode).JSON(data)
}

func SendError(c *fiber.Ctx, statusCode int, message string) error {
    return c.Status(statusCode).JSON(fiber.Map{"error": message})
}
