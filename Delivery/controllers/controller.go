package controllers

import (
	"context"
	"net/http"
	domain "task_manager/Domain"
	infrastructure "task_manager/Infrastructure"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController struct {
	userUsecase domain.UserUsecase
}

type TaskController struct {
	taskUsecase domain.TaskUsecase
}

// Constructor for TaskController
func NewUserController(userUsecase domain.UserUsecase) *UserController {
	return &UserController{userUsecase: userUsecase}
}

func NewTaskController(taskUsecase domain.TaskUsecase) *TaskController {
	return &TaskController{taskUsecase: taskUsecase}
}

// ------------------------- User Handlers -------------------------

func (userControl *UserController) Register(c *gin.Context) {
	var req domain.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	insertResult, err := userControl.userUsecase.Register(ctx, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	insertedID := insertResult.InsertedID.(primitive.ObjectID)

	// Registration successful
	// Return user information with the hashed password
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       insertedID.Hex(),
			"username": req.Username,
			"role":     domain.RoleUser,
		}})
}

// Handles user authentication
func (userControl *UserController) Login(c *gin.Context) {
	var req domain.LoginRequest

	// Bind the request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the user by username in the database
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tokenString, err := userControl.userUsecase.Login(ctx, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()}) // 401 Unauthorized
		return
	}

	// Return the token to the client
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// ------------------------- Task Handlers -------------------------

func (taskControl *TaskController) GetAllTask(c *gin.Context) {

	ctx := c.Request.Context()

	tasks, err := taskControl.taskUsecase.GetAllTask(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// Get specific task based on ID.
func (taskControl *TaskController) GetTaskByID(c *gin.Context) {
	id := c.Param("id")
	// Use request context
	ctx := c.Request.Context()

	task, err := taskControl.taskUsecase.GetTaskByID(ctx, id)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, task)
}

// Update existing task.
func (taskControl *TaskController) UpdateTask(c *gin.Context) {
	id := c.Param("id")

	var task domain.Task

	// Bind the request data to the variable created.
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Request Context
	ctx := c.Request.Context()

	err := taskControl.taskUsecase.UpdateTask(ctx, id, task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated"})
}

// Delete exiting task.
func (taskControl *TaskController) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	// Request Context
	ctx := c.Request.Context()

	err := taskControl.taskUsecase.DeleteTask(ctx, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// Create new task.
func (taskControl *TaskController) NewTask(c *gin.Context) {
	var newTask domain.Task

	// Bind request to variable
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the authenticated username from the contetx
	username, _, err := infrastructure.GetUserFromContext(c) // Use helper function
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newTask.CreatedBy = username

	// Request Context
	ctx := c.Request.Context()

	insertResult, err := taskControl.taskUsecase.NewTask(ctx, newTask)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the inserted ID
	insertedID, _ := insertResult.InsertedID.(primitive.ObjectID)

	newTask.ID = insertedID

	c.JSON(http.StatusCreated, newTask)
}
