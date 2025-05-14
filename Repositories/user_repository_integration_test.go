// Repositories/user_repository_integration_test.go
package repositories_test

import (
	"context"
	"errors"
	domain "task_manager/Domain"
	repositories "task_manager/Repositories"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const testUserCollectionName = "users_integration_test_coll"

// Helper to get a clean user collection for each test
func getUserTestCollection(t *testing.T) *mongo.Collection {
	require.NotNil(t, testDBClient, "Database client not initialized. TestMain setup might have failed.")
	collection := testDBClient.Database(TestDatabaseName).Collection(testUserCollectionName)
	// Clean the collection before each test that uses it
	_, err := collection.DeleteMany(context.Background(), bson.M{})
	require.NoError(t, err, "Failed to clean user test collection")
	return collection
}

func TestUserRepository_Integration(t *testing.T) {
	// Ensure TestMain has run and testDBClient is initialized
	if testDBClient == nil {
		t.Fatal("testDBClient is nil. TestMain setup for DB connection likely failed or was skipped.")
	}

	// Create repository instance using the testDBClient and specific test collection name
	userRepo := repositories.NewUserRepository(testDBClient, TestDatabaseName, testUserCollectionName)
	require.NotNil(t, userRepo, "NewUserRepository returned nil")

	ctx := context.Background() // Use a fresh context for operations

	t.Run("CreateUser_Success", func(t *testing.T) {
		_ = getUserTestCollection(t) // Clean collection

		userToCreate := &domain.User{
			Username:     "integ_testuser1",
			PasswordHash: "hashedpassword1",
			Role:         domain.RoleUser,
		}

		insertResult, err := userRepo.CreateUser(ctx, userToCreate)
		require.NoError(t, err, "CreateUser should not return an error on success")
		require.NotNil(t, insertResult, "InsertResult should not be nil")
		require.NotNil(t, insertResult.InsertedID, "InsertedID should not be nil")
		insertedID, ok := insertResult.InsertedID.(primitive.ObjectID)
		require.True(t, ok, "InsertedID should be a primitive.ObjectID")

		// Verify directly in the database
		var foundUser domain.User
		collection := testDBClient.Database(TestDatabaseName).Collection(testUserCollectionName)
		err = collection.FindOne(ctx, bson.M{"_id": insertedID}).Decode(&foundUser)
		require.NoError(t, err, "Failed to find the created user directly in DB")
		assert.Equal(t, userToCreate.Username, foundUser.Username)
		assert.Equal(t, userToCreate.PasswordHash, foundUser.PasswordHash)
		assert.Equal(t, userToCreate.Role, foundUser.Role)
		assert.Equal(t, insertedID, foundUser.ID, "Stored ID should match inserted ID")
	})

	t.Run("CreateUser_Failure_UsernameAlreadyExists", func(t *testing.T) {
		_ = getUserTestCollection(t) // Clean collection

		user1 := &domain.User{Username: "duplicate_integ_user", PasswordHash: "hash1", Role: domain.RoleUser}
		_, err := userRepo.CreateUser(ctx, user1) // Insert the first user
		require.NoError(t, err, "First CreateUser call failed unexpectedly")

		user2 := &domain.User{Username: "duplicate_integ_user", PasswordHash: "hash2", Role: domain.RoleUser}
		insertResult, err := userRepo.CreateUser(ctx, user2) // Attempt to insert user with the same username

		require.Error(t, err, "CreateUser should return an error for duplicate username")
		assert.Nil(t, insertResult, "InsertResult should be nil on error")
		// Your repository already checks for this and returns a specific error
		assert.EqualError(t, err, "username is already taken")

		// Verify only one user with that username exists in the DB
		collection := testDBClient.Database(TestDatabaseName).Collection(testUserCollectionName)
		count, err := collection.CountDocuments(ctx, bson.M{"username": "duplicate_integ_user"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "Should only be one user with that username in DB")
	})

	t.Run("FindUserByUsername_Success_UserExists", func(t *testing.T) {
		userCollection := getUserTestCollection(t) // Clean and get collection

		expectedUser := &domain.User{
			ID:           primitive.NewObjectID(), // Pre-assign ID for exact match if desired
			Username:     "findme_integ",
			PasswordHash: "findmehash_integ",
			Role:         domain.RoleAdmin,
		}
		_, err := userCollection.InsertOne(ctx, expectedUser) // Manually insert for the test
		require.NoError(t, err, "Failed to insert test user directly into DB")

		foundUser, err := userRepo.FindUserByUsername(ctx, "findme_integ")
		require.NoError(t, err, "FindUserByUsername should not return an error for existing user")
		require.NotNil(t, foundUser, "Found user should not be nil")

		assert.Equal(t, expectedUser.ID, foundUser.ID)
		assert.Equal(t, expectedUser.Username, foundUser.Username)
		assert.Equal(t, expectedUser.PasswordHash, foundUser.PasswordHash)
		assert.Equal(t, expectedUser.Role, foundUser.Role)
	})

	t.Run("FindUserByUsername_Failure_UserDoesNotExist", func(t *testing.T) {
		_ = getUserTestCollection(t) // Clean collection

		foundUser, err := userRepo.FindUserByUsername(ctx, "idonotexist_integ")
		require.Error(t, err, "FindUserByUsername should return an error for non-existent user")
		assert.Nil(t, foundUser, "Found user should be nil when not found")
		// Your repository wraps mongo.ErrNoDocuments
		assert.EqualError(t, err, "user not found")
	})

	t.Run("FindUserByUsername_ContextTimeout", func(t *testing.T) {
		_ = getUserTestCollection(t) // Clean collection
		// This test is harder to make reliable as it depends on the DB being slow.
		// The repository itself has a 5s timeout on FindUserByUsername.
		// We can test that a short-lived context passed from outside also works.

		userToInsert := &domain.User{Username: "timeout_user_integ", PasswordHash: "hash"}
		_, err := userRepo.CreateUser(ctx, userToInsert) // Use the main ctx for setup
		require.NoError(t, err)

		shortCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond) // Extremely short timeout
		defer cancel()

		_, err = userRepo.FindUserByUsername(shortCtx, "timeout_user_integ")
		require.Error(t, err, "FindUserByUsername should error out with a cancelled context")
		// The error could be context.DeadlineExceeded from the shortCtx,
		// or if the repo's internal timeout hits first, it might be that.
		// The repo's FindOne itself will return an error if its own timeout (5s) or the passed context's timeout is exceeded.
		assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || mongo.IsTimeout(err),
			"Expected context deadline/canceled error or mongo timeout, got: %v", err)
	})
}
