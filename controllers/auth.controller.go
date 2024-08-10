package controllers

import (
    "context"
    "log"
    "regexp"
    "strings"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/manlikehenryy/go-task-manager/helpers"
    "github.com/manlikehenryy/go-task-manager/models"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

func validateEmail(email string) bool {
    emailPattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
    re := regexp.MustCompile(emailPattern)
    return re.MatchString(email)
}



func Register(c *fiber.Ctx) error {
    var data map[string]interface{}
    if err := c.BodyParser(&data); err != nil {
        log.Println("Unable to parse body:", err)
        return helpers.SendError(c, fiber.StatusBadRequest, "Invalid request payload")
    }

    password, passwordOk := data["password"].(string)
    if !passwordOk || len(password) <= 6 {
        return helpers.SendError(c, fiber.StatusBadRequest, "Password must be greater than 6 characters")
    }

    email, emailOk := data["email"].(string)
    if !emailOk || !validateEmail(strings.TrimSpace(email)) {
        return helpers.SendError(c, fiber.StatusBadRequest, "Invalid email address")
    }
    
    // Check if the email already exists
    var user models.User
    err := usersCollection.FindOne(context.Background(), bson.M{"email": strings.TrimSpace(email)}).Decode(&user)
    if err != mongo.ErrNoDocuments {
        if err != nil {
            log.Println("Database error:", err)
            return helpers.SendError(c, fiber.StatusInternalServerError, "Failed to create account")
        }
        return helpers.SendError(c, fiber.StatusBadRequest, "Email already exists")
    }

    user = models.User{
        FirstName: data["firstName"].(string),
        LastName:  data["lastName"].(string),
        Phone:     data["phone"].(string),
        Email:     strings.TrimSpace(email),
    }

    user.SetPassword(password)
    insertResult, err := usersCollection.InsertOne(context.Background(), user)
    if err != nil {
        log.Println("Database error:", err)
        return helpers.SendError(c, fiber.StatusInternalServerError, "Failed to create account")
    }

    user.ID = insertResult.InsertedID.(primitive.ObjectID)

    return helpers.SendJSON(c, fiber.StatusCreated, fiber.Map{
        "data":    user,
        "message": "Account created successfully",
    })
}

func Login(c *fiber.Ctx) error {
    var data map[string]string
    if err := c.BodyParser(&data); err != nil {
        log.Println("Unable to parse body:", err)
        return helpers.SendError(c, fiber.StatusBadRequest, "Invalid request payload")
    }

    email, emailOk := data["email"]
    password, passwordOk := data["password"]
    if !emailOk || !passwordOk {
        return helpers.SendError(c, fiber.StatusBadRequest, "Email and password are required")
    }

    filter := bson.M{"email": strings.TrimSpace(email)}

    var user models.User
    err := usersCollection.FindOne(context.Background(), filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return helpers.SendError(c, fiber.StatusUnauthorized, "Incorrect email address or password")
        }
        log.Println("Database error:", err)
        return helpers.SendError(c, fiber.StatusInternalServerError, "Database error")
    }

    if err := user.ComparePassword(password); err != nil {
        return helpers.SendError(c, fiber.StatusUnauthorized, "Incorrect email address or password")
    }

    token, err := helpers.GenerateJwt(user.ID.Hex())
    if err != nil {
        log.Println("Token generation error:", err)
        return helpers.SendError(c, fiber.StatusInternalServerError, "Failed to generate token")
    }

    cookie := fiber.Cookie{
        Name:     "jwt",
        Value:    token,
        Expires:  time.Now().Add(time.Hour * 24),
        HTTPOnly: true,
    }
    c.Cookie(&cookie)

    return helpers.SendJSON(c, fiber.StatusOK, fiber.Map{
        "data":    user,
        "message": "Logged in successfully",
    })
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return helpers.SendJSON(c, fiber.StatusOK, fiber.Map{
		"message": "Logged out successfully",
	})
}
