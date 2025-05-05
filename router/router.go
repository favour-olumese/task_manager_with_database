package router

import (
	"net/http"
	"task_manager/controllers"
	"task_manager/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(taskController *controllers.TaskController, userController *controllers.UserController) *gin.Engine {
	router := gin.Default()

	// Public routes (no authentication required)
	router.GET("", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Welcome to Favour Olumese Task Manager"})
	})

	// User authentication routes (public)
	userGroup := router.Group("/users") // Group related user routes
	{
		userGroup.POST("/register", userController.Register)
		userGroup.POST("/login", userController.Login)
	}

	// Protect tasks routes (authenication required)
	// Apply the AuthRequired middleware to this group
	protectedTaskGroup := router.Group("/tasks")
	protectedTaskGroup.Use(middleware.AuthRequired()) // Apply authentication middleware to all routes in this group
	{
		protectedTaskGroup.GET("", taskController.GetAllTask)
		protectedTaskGroup.GET("/:id", taskController.GetTaskByID)
		protectedTaskGroup.PUT("/:id", taskController.UpdateTask)
		protectedTaskGroup.DELETE("/:id", taskController.DeleteTask)
		protectedTaskGroup.POST("", taskController.NewTask)
	}
	return router
}
