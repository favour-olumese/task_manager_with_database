package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"task_manager/auth"
	"task_manager/data"
	"task_manager/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TASK CONTROLLERS

// Holds the MongoDB collection dependency
type TaskController struct {
	TaskCollection *mongo.Collection
}

// Create a new instance of TaskController
func NewTaskController(collection *mongo.Collection) *TaskController {
	return &TaskController{
		TaskCollection: collection,
	}
}

// Get all tasks.
func (tc *TaskController) GetAllTask(c *gin.Context) { // Method on TaskController

	var tasks []models.Task

	ctx := c.Request.Context()

	cursor, err := tc.TaskCollection.Find(ctx, bson.D{{}})

	if err != nil {
		log.Printf("Error finding tasks: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	// Close the cursor when done.
	defer cursor.Close(ctx)

	// Finding multiple documents returns a cursor.
	// Iterating through the cursor.
	for cursor.Next(ctx) {
		var element models.Task

		err := cursor.Decode(&element)

		if err != nil {
			log.Printf("Error decoding task: %v", err)

			continue // Go to the next element.
		}

		tasks = append(tasks, element)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to process task cursor"})
		return
	}

	// Response
	c.JSON(http.StatusOK, tasks)
}

// Get specific task based on ID.
func (tc *TaskController) GetTaskByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var findTask models.Task

	// Use request context
	ctx := c.Request.Context()

	filter := bson.M{"_id": id}

	err = tc.TaskCollection.FindOne(ctx, filter).Decode(&findTask)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Task not found for ID: %s", idStr)
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		} else {
			log.Printf("Error finding task %s %v", idStr, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task"})
		}
		return
	}

	c.JSON(http.StatusOK, findTask)
}

// Update existing task.
func (tc *TaskController) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var updatedTask models.Task

	// Bind the request data to the variable created.
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setFields := bson.M{}

	if updatedTask.Title != "" {
		setFields["title"] = updatedTask.Title
	}

	// Description
	if updatedTask.Description != "" {
		setFields["description"] = updatedTask.Description
	}

	// Status
	if updatedTask.Status != "" {
		setFields["status"] = updatedTask.Status
	}

	// Due Date
	if !updatedTask.DueDate.IsZero() {
		setFields["due_date"] = updatedTask.DueDate
	}

	// Confirm that the fields are not empty
	if len(setFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No field provided for update"})
		return
	}

	updatingTask := bson.M{"$set": setFields}

	filter := bson.M{"_id": id}

	ctx := c.Request.Context()

	result, err := tc.TaskCollection.UpdateOne(ctx, filter, updatingTask)

	if err != nil {
		log.Printf("Error updating task %s: %v", idStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	if result.MatchedCount == 0 {
		log.Printf("Task not found for update: %s", idStr)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated"})
}

// Delete exiting task.
func (tc *TaskController) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	filter := bson.M{"_id": id}

	// Request Context
	ctx := c.Request.Context()

	result, err := tc.TaskCollection.DeleteOne(ctx, filter)

	if err != nil {
		log.Printf("Error deleting task %s from MongoDB: %v\n", idStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	// Check if task to be deleted exists.
	if result.DeletedCount == 0 {
		log.Printf("No task found for the given id %s", idStr)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// Create new task.
func (tc *TaskController) NewTask(c *gin.Context) {
	var newTask models.Task

	// Bind request to variable
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Client side ID generation.
	// newTask.ID = primitive.NewObjectID()

	// Request Context
	ctx := c.Request.Context()

	insertResult, err := tc.TaskCollection.InsertOne(ctx, newTask)

	if err != nil {
		log.Printf("Error inserting new task into the database: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during insert"})
		return
	}

	// Get the inserted ID
	insertedID := insertResult.InsertedID.(primitive.ObjectID)
	newTask.ID = insertedID

	c.JSON(http.StatusCreated, newTask)
}

// USER CONTROLLER

// Hold the dependencies for user-related operations
type UserController struct {
	UserCollection *mongo.Collection
}

// Create a new instance of UserController
func NewUserController(collection *mongo.Collection) *UserController {
	return &UserController{
		UserCollection: collection,
	}
}

// Handles user's registration
func (uc *UserController) Register(c *gin.Context) {
	var req models.ReisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	newUser := models.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Role:         models.RoleUser,
	}

	// Create user in the database
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	insertResult, err := data.CreateUser(ctx, uc.UserCollection, &newUser)
	if err != nil {
		// Check if username has alread been taken
		if err.Error() == fmt.Sprintf("username '%s' is already taken", newUser.Username) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // 409 Conflict
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"}) // 500 Internal Server Error
		}
		return
	}

	insertedID, ok := insertResult.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Printf("Warning: InsertedID is not an ObjetID, it's %T: %v", insertResult.InsertedID, insertResult.InsertedID)
	} else {
		newUser.ID = insertedID
	}

	// Registration successful
	// Return user information with the hashed password
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       newUser.ID.Hex(),
			"username": newUser.Username,
			"role":     newUser.Role,
		}})
}

// Handles user authentication
func (uc *UserController) Login(c *gin.Context) {
	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Bind the request body
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the user by username in the database
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	user, err := data.FindUserByUsername(ctx, uc.UserCollection, loginRequest.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"}) // 401 Unauthorized
		return
	}

	// Compare the provided password with the stired hashed password
	err = auth.ComparePasswords(user.PasswordHash, loginRequest.Password)
	if err != nil {
		// Password do not match
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenString, err := auth.GenerateToken(user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return the token to the client
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
