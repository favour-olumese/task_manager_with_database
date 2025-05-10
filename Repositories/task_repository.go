package repositories

import (
	"context"
	"errors"
	domain "task_manager/Domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Defines the interface
type TaskRepository interface {
	GetAllTask(ctx context.Context) ([]domain.Task, error)
	GetTaskByID(ctx context.Context, id string) (domain.Task, error)
	UpdateTask(ctx context.Context, id string, updatedTask domain.Task) error
	DeleteTask(ctx context.Context, id string) error
	NewTask(ctx context.Context, task domain.Task) (*mongo.InsertOneResult, error)
}

type taskRepository struct {
	collection *mongo.Collection
}

// Ensure *taskRepostory implements TaskRepository
var _ TaskRepository = (*taskRepository)(nil)

func NewTaskRepository(db *mongo.Client, dbName, collectionName string) TaskRepository {
	return &taskRepository{
		collection: db.Database(dbName).Collection(collectionName),
	}
}

// Gets all Tasks
func (repo *taskRepository) GetAllTask(ctx context.Context) ([]domain.Task, error) {
	var tasks []domain.Task

	cursor, err := repo.collection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	// Close the cursor when done.
	defer cursor.Close(ctx)

	// Finding multiple documents returns a cursor.
	// Iterating through the cursor.
	for cursor.Next(ctx) {
		var element domain.Task

		err := cursor.Decode(&element)

		if err != nil {
			return nil, err
		}

		tasks = append(tasks, element)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// Gets task by ID
func (repo *taskRepository) GetTaskByID(ctx context.Context, id string) (domain.Task, error) {
	var findTask domain.Task

	// Convert string id to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.Task{}, errors.New("invalid task ID format")
	}

	filter := bson.M{"_id": objectID}

	err = repo.collection.FindOne(ctx, filter).Decode(&findTask)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.Task{}, errors.New("task not found")
		}
		return domain.Task{}, err
	}
	return findTask, nil
}

// Update and existing task
func (repo *taskRepository) UpdateTask(ctx context.Context, id string, updatedTask domain.Task) error {
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
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
		return errors.New("no field provided")
	}

	updatingTask := bson.M{"$set": setFields}

	filter := bson.M{"_id": objectID}

	result, err := repo.collection.UpdateOne(ctx, filter, updatingTask)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("task not found")
	}

	return nil
}

func (repo *taskRepository) DeleteTask(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}

	result, err := repo.collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	// Check if task to be deleted exists.
	if result.DeletedCount == 0 {
		return errors.New("task not found")
	}

	return nil
}

// Creates a new task
func (repo *taskRepository) NewTask(ctx context.Context, task domain.Task) (*mongo.InsertOneResult, error) {
	result, err := repo.collection.InsertOne(ctx, task)

	return result, err
}
