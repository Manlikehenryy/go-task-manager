package helpers

import (
	"context"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/manlikehenryy/go-task-manager/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CheckEmailExists(c *fiber.Ctx, usersCollection *mongo.Collection, email string) error {
    filter := bson.M{"email": strings.TrimSpace(email)}

    var user models.User

    err := usersCollection.FindOne(context.Background(), filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            // No document found, so email does not exist
            return nil
        }
        log.Println("Database error:", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Database error",
        })
    }

    // If no error and we got a result, the email exists
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
        "error": "Email already exists",
    })
}
