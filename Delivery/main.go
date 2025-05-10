package main

import (
	"context"
	"log"
	"task_manager/Delivery/router"
	infrastructure "task_manager/Infrastructure"
	"time"
)

func main() {
	// Set up context with a timeout for database connection
	dbConnectContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to Database.
	dbClient, err := infrastructure.ConnectDB(dbConnectContext)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Disconnect database when main exits.
	defer infrastructure.DisconnectDB(dbClient)

	routes := router.SetupRouter(dbClient)

	// Start server
	log.Println("Starting server on port 8080.")
	if err := routes.Run("localhost:8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
