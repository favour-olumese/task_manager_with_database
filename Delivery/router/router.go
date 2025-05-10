package router

import (
	"net/http"
	"task_manager/Delivery/controllers"
	infrastructure "task_manager/Infrastructure"
	repositories "task_manager/Repositories"
	usecases "task_manager/Usecases"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRouter(dbClient *mongo.Client) *gin.Engine {
	// Initialize repositories
	taskRepo := repositories.NewTaskRepository(dbClient, "task_manager", "tasks")
	userRepo := repositories.NewUserRepository(dbClient, "task_manager", "user")

	// Initialize services
	jwtService := infrastructure.NewJWTService()
	passwordService := infrastructure.NewPasswordService()
	authMiddleware := infrastructure.NewAuthMiddleware(jwtService)

	// Initialize usecases
	taskUsecase := usecases.NewTaskUsecase(taskRepo)
	userUsecase := usecases.NewUserUsecase(userRepo, passwordService, jwtService)

	taskController := controllers.NewTaskController(taskUsecase)
	userController := controllers.NewUserController(userUsecase)

	// Setup Gin router
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
	protectedTaskGroup.Use(authMiddleware.AuthRequired()) // Apply authentication middleware to all routes in this group
	{
		protectedTaskGroup.GET("", taskController.GetAllTask)
		protectedTaskGroup.GET("/:id", taskController.GetTaskByID)
		protectedTaskGroup.PUT("/:id", taskController.UpdateTask)
		protectedTaskGroup.DELETE("/:id", taskController.DeleteTask)
		protectedTaskGroup.POST("", taskController.NewTask)
	}
	return router
}
