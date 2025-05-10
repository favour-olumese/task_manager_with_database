package repositories

import (
	"context"
	"errors"
	domain "task_manager/Domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepository struct {
	collection *mongo.Collection
}

var _ domain.UserRepository = (*userRepository)(nil)

func NewUserRepository(db *mongo.Client, dbName, collectionName string) domain.UserRepository {
	return &userRepository{
		collection: db.Database(dbName).Collection(collectionName),
	}
}

// Creates a new user
func (repo *userRepository) CreateUser(ctx context.Context, user *domain.User) (*mongo.InsertOneResult, error) {

	// Check if user already exists.
	existingUser := domain.User{}

	err := repo.collection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)

	if err == nil {
		// User found.
		return nil, errors.New("username is already taken")
	}

	// If no document is returned, it means no user currently have that username.
	if err != mongo.ErrNoDocuments {
		return nil, errors.New("database error while checking username")
	}

	result, err := repo.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Get a user by their username.
func (repo *userRepository) FindUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User

	filter := bson.M{"username": username}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := repo.collection.FindOne(ctx, filter).Decode(&user)

	// User does not exist
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}
