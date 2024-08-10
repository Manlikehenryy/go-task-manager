package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manlikehenryy/go-task-manager/controllers"
	"github.com/manlikehenryy/go-task-manager/database"
	"github.com/manlikehenryy/go-task-manager/middleware"
)

func Setup(app *fiber.App) {
    controllers.InitDB(database.Client.Database("go_task_manager"))

	app.Post("/api/register", controllers.Register)
	app.Post("/api/login", controllers.Login)
	app.Get("/api/logout", controllers.Logout)

	app.Use(middleware.IsAuthenticated)

	app.Post("/api/task", controllers.CreateTask)
	app.Get("/api/task", controllers.GetAllTasks)
	app.Get("/api/task/:id", controllers.GetTask)
	app.Put("/api/task/:id", controllers.UpdateTask)
	app.Get("/api/user-tasks", controllers.UsersTask)
	app.Delete("/api/task/:id", controllers.DeleteTask)
}