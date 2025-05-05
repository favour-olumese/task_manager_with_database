package data

import (
	"context"
	"fmt"
	"task_manager/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Get user collection from the database client
func NewUserCollection(dbClient *mongo.Client, dbName string, collectionName string) *mongo.Collection {
	collection := dbClient.Database(dbName).Collection(collectionName)
	return collection
}

// Insert a new user into the database.
func CreateUser(ctx context.Context, collection *mongo.Collection, user *models.User) (*mongo.InsertOneResult, error) {

	// Check if user already exists.
	existingUser := models.User{}

	err := collection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
	if err == nil {
		// User found.
		return nil, fmt.Errorf("username '%s' is already taken", user.Username)
	}

	// If not document is returned, it means no user currently have that username.
	if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("database error while checking username: %w", err)
	}

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new user: %w", err)
	}

	return result, nil
}

// Get a user by their username.
func FindUserByUsername(ctx context.Context, collection *mongo.Collection, username string) (*models.User, error) {
	var user models.User

	filter := bson.M{"username": username}
	err := collection.FindOne(ctx, filter).Decode(&user)

	// User does not exist
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("database error finding user: %w", err)
	}

	return &user, nil
}
