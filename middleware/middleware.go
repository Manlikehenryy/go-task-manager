package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manlikehenryy/go-task-manager/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func IsAuthenticated(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

    // Parse the JWT token from the cookie
    userIdStr, err := helpers.ParseJwt(cookie)
    if err != nil {
        c.Status(fiber.StatusUnauthorized)
        return c.JSON(fiber.Map{
            "message": "Unauthorized",
        })
    }

    // Convert the user ID string to primitive.ObjectID
    userId, err := primitive.ObjectIDFromHex(userIdStr)
    if err != nil {
        c.Status(fiber.StatusUnauthorized)
        return c.JSON(fiber.Map{
            "message": "Invalid user ID",
        })
    }

    // Store the user ID in the request context
    c.Locals("userId", userId)

    return c.Next()
}