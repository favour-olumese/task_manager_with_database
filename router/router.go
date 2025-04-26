package router

import (
	"net/http"
	"task_manager/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(taskController *controllers.TaskController) *gin.Engine {
	router := gin.Default()

	router.GET("", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Welcome to Favour Olumese Task Manager"})
	})
	router.GET("/tasks", taskController.GetAllTask)
	router.GET("/tasks/:id", taskController.GetTaskByID)
	router.PUT("/tasks/:id", taskController.UpdateTask)
	router.DELETE("/tasks/:id", taskController.DeleteTask)
	router.POST("/tasks", taskController.NewTask)

	return router
}
