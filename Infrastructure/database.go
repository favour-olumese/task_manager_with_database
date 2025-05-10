package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Establish a connection to MongoDB and returns the client.
// Returns an error if connection fails.
func ConnectDB(ctx context.Context) (*mongo.Client, error) {

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Check the connection.
	err = client.Ping(ctx, nil)

	if err != nil {
		// Disconnect if ping fails
		_ = client.Disconnect(context.TODO())

		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// collection := client.Database("task_manager").Collection("tasks")
	log.Println("Successfully connected and pinged MongoDB.")
	return client, nil // Return client on success
}

// Close the Database Connecction.
func DisconnectDB(client *mongo.Client) {
	if client == nil {
		return
	}

	disconnectContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err := client.Disconnect(disconnectContext)

	if err != nil {
		log.Printf("failed to disconnect MongoDB: %v", err)
	} else {
		log.Println("Connection to MongoDB closed.")
	}
}
