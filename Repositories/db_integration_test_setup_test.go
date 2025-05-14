package repositories_test // Must be _test package

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	testDBClient     *mongo.Client
	TestDatabaseName = "task_manager_integration_test_db" // Dedicated test DB name
)

// TestMain sets up the DB connection before running tests in this package and disconnects after.
func TestMain(m *testing.M) {
	log.Println("Setting up MongoDB connection for integration tests...")

	// Get MongoDB URI from environment variable or use a default
	mongoURI := os.Getenv("MONGO_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // Default for local testing
		log.Printf("MONGO_TEST_URI not set, using default: %s\n", mongoURI)
	} else {
		log.Printf("Using MONGO_TEST_URI: %s\n", mongoURI)
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second) // Increased timeout for CI
	defer cancel()

	var err error
	testDBClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB for testing at %s: %v", mongoURI, err)
	}

	err = testDBClient.Ping(ctx, nil)
	if err != nil {
		// Attempt to disconnect even if ping fails
		_ = testDBClient.Disconnect(context.Background())
		log.Fatalf("Failed to ping MongoDB for testing: %v", err)
	}
	log.Println("Successfully connected and pinged MongoDB for integration tests.")

	// --- Run the tests ---
	exitCode := m.Run()
	// --- End of tests ---

	// Teardown: Clean up the test database and disconnect
	log.Printf("Cleaning up test database: %s...\n", TestDatabaseName)
	// It's crucial to use a context for DB operations, even in teardown.
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()

	if err := testDBClient.Database(TestDatabaseName).Drop(cleanupCtx); err != nil {
		log.Printf("Failed to drop test database '%s': %v", TestDatabaseName, err)
	} else {
		log.Printf("Successfully dropped test database '%s'.\n", TestDatabaseName)
	}

	disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer disconnectCancel()
	if err := testDBClient.Disconnect(disconnectCtx); err != nil {
		log.Fatalf("Failed to disconnect from MongoDB after testing: %v", err)
	}
	log.Println("Disconnected from MongoDB.")
	os.Exit(exitCode)
}
