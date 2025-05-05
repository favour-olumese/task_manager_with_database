package main

import (
	"context"
	"fmt"
	"log"
	"task_manager/controllers"
	"task_manager/data"
	"task_manager/router"
	"time"
)

func main() {
	// Set up context with a timeout for database connection
	dbConnectContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	// Connect to Database.
	dbClient, err := data.ConnectDB(dbConnectContext)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Disconnect database when main exits.
	defer data.DisconnectDB(dbClient)

	// Initialize Task and User Collections
	taskCollection := data.NewTaskCollection(dbClient, "task_manager", "tasks")
	userCollection := data.NewTaskCollection(dbClient, "task_manager", "users")

	// Initialize Controllers
	taskController := controllers.NewTaskController(taskCollection)
	userController := controllers.NewUserController(userCollection)

	fmt.Printf("Task collection has been created: %v\n", taskCollection)
	fmt.Printf("User collection has been created: %v\n", userCollection)

	routes := router.SetupRouter(taskController, userController)

	log.Println("Starting server on port 8080.")
	if err := routes.Run("localhost:8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
