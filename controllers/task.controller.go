package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/manlikehenryy/go-task-manager/helpers"
	"github.com/manlikehenryy/go-task-manager/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateTask(c *fiber.Ctx) error {
	var task models.Task

	if err := c.BodyParser(&task); err != nil {

		fmt.Println("Unable to parse body:", err)

		return helpers.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	userId, ok := c.Locals("userId").(primitive.ObjectID)
	if !ok {
		return helpers.SendError(c, fiber.StatusUnauthorized, "User ID not found in context")
	}

	task.UserId = userId

	insertResult, err := tasksCollection.InsertOne(context.Background(), task)
	if err != nil {
		log.Println("Database error:", err)
		return helpers.SendError(c, fiber.StatusInternalServerError, "Failed to create account")
	}

	task.ID = insertResult.InsertedID.(primitive.ObjectID)

	return helpers.SendJSON(c, fiber.StatusCreated, fiber.Map{
		"data":    task,
		"message": "Post created successfully",
	})
}

func GetAllTasks(c *fiber.Ctx) error {

	// Get pagination parameters from query string
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("perPage", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	if page == 0 {
		page = 1
		limit = math.MaxInt
	}

	offset := (page - 1) * limit

	total, err := tasksCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to count posts")
	}

	cursor, err := tasksCollection.Find(context.Background(), bson.M{}, options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve posts")
	}
	defer cursor.Close(context.Background())

	var tasks []models.Task
	for cursor.Next(context.Background()) {
		var task models.Task
		if err := cursor.Decode(&task); err != nil {
			log.Println("Database error:", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to decode posts")
		}
		tasks = append(tasks, task)
	}

	pageCount := int(math.Ceil(float64(total) / float64(limit)))
	hasNextPage := page < pageCount
	hasPrevPage := page > 1

	if limit == math.MaxInt {
		limit = int(total)
	}

	return c.JSON(fiber.Map{
		"data":    tasks,
		"message": "Posts fetched successfully",
		"meta": fiber.Map{
			"page":      page,
			"perPage":   limit,
			"total":     total,
			"pageCount": pageCount,
			"nextPage": func() int {
				if hasNextPage {
					return page + 1
				}
				return 0
			}(),
			"hasNextPage": hasNextPage,
			"hasPrevPage": hasPrevPage,
		},
	})
}

func GetTask(c *fiber.Ctx) error {

	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return helpers.SendError(c, fiber.StatusBadRequest, "Invalid task ID")
	}

	userId, ok := c.Locals("userId").(primitive.ObjectID)
	if !ok {
		return helpers.SendError(c, fiber.StatusUnauthorized, "User ID not found in context")
	}

	var task models.Task
	err = tasksCollection.FindOne(context.Background(), bson.M{"_id": id, "userId": userId}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return helpers.SendError(c, fiber.StatusNotFound, "Task not found")
		}
		return helpers.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve task")
	}

	return helpers.SendJSON(c, fiber.StatusOK, fiber.Map{
		"data": task,
	})
}

func UpdateTask(c *fiber.Ctx) error {

	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return helpers.SendError(c, fiber.StatusBadRequest, "Invalid task ID")
	}

	userId, ok := c.Locals("userId").(primitive.ObjectID)
	if !ok {
		return helpers.SendError(c, fiber.StatusUnauthorized, "User ID not found in context")
	}

	var task models.Task
	if err := c.BodyParser(&task); err != nil {
		return helpers.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	var existingTask models.Task
	err = tasksCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&existingTask)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return helpers.SendError(c, fiber.StatusNotFound, "Task not found")
		}
		return helpers.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve task")
	}

	if existingTask.UserId != userId {
		return helpers.SendError(c, fiber.StatusForbidden, "Unauthorized to update this post")
	}

	update := bson.M{
		"$set": bson.M{
			"title":  task.Title,
			"desc":   task.Desc,
			"status": task.Status,
		},
	}

	result, err := tasksCollection.UpdateOne(context.Background(), bson.M{"_id": id, "userId": userId}, update)
	if err != nil {
		log.Println("Database error:", err)
		return helpers.SendError(c, fiber.StatusInternalServerError, "Failed to update post")
	}

	if result.MatchedCount == 0 {
		return helpers.SendError(c, fiber.StatusNotFound, "Task not found")
	}

	return helpers.SendJSON(c, fiber.StatusOK, fiber.Map{
		"message": "Post updated successfully",
	})
}

func UsersTask(c *fiber.Ctx) error {

	userId, ok := c.Locals("userId").(primitive.ObjectID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "User ID not found in context")
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("perPage", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	if page == 0 {
		page = 1
	}

	offset := (page - 1) * limit

	total, err := tasksCollection.CountDocuments(context.Background(), bson.M{"userId": userId})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to count tasks")
	}

	cursor, err := tasksCollection.Find(
		context.Background(),
		bson.M{"userId": userId},
		options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)),
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve tasks")
	}
	defer cursor.Close(context.Background())

	var tasks []models.Task
	for cursor.Next(context.Background()) {
		var task models.Task
		if err := cursor.Decode(&task); err != nil {
			log.Println("Database error:", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to decode tasks")
		}
		tasks = append(tasks, task)
	}

	pageCount := int(math.Ceil(float64(total) / float64(limit)))
	hasNextPage := page < pageCount
	hasPrevPage := page > 1

	if limit == math.MaxInt {
		limit = int(total)
	}

	return c.JSON(fiber.Map{
		"data":    tasks,
		"message": "Tasks fetched successfully",
		"meta": fiber.Map{
			"page":      page,
			"perPage":   limit,
			"total":     total,
			"pageCount": pageCount,
			"nextPage": func() int {
				if hasNextPage {
					return page + 1
				}
				return 0
			}(),
			"hasNextPage": hasNextPage,
			"hasPrevPage": hasPrevPage,
		},
	})
}

func DeleteTask(c *fiber.Ctx) error {

	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return helpers.SendError(c, fiber.StatusBadRequest, "Invalid task ID")
	}

	userId, ok := c.Locals("userId").(primitive.ObjectID)
	if !ok {
		return helpers.SendError(c, fiber.StatusUnauthorized, "User ID not found in context")
	}

	var existingTask models.Task
	err = tasksCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&existingTask)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return helpers.SendError(c, fiber.StatusNotFound, "Task not found")
		}
		return helpers.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve task")
	}

	if existingTask.UserId != userId {
		return helpers.SendError(c, fiber.StatusForbidden, "Unauthorized to delete this task")
	}

	filter := bson.M{"_id": id, "userId": userId}
	result, err := tasksCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return helpers.SendError(c, fiber.StatusInternalServerError, "Failed to delete task")
	}

	if result.DeletedCount == 0 {
		return helpers.SendError(c, fiber.StatusNotFound, "Task not found or unauthorized")
	}

	return helpers.SendJSON(c, fiber.StatusOK, fiber.Map{
		"message": "Task deleted successfully",
	})
}
