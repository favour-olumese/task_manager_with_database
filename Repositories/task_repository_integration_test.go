// Repositories/task_repository_integration_test.go
package repositories_test

import (
	"context"
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

const testTaskCollectionName = "tasks_integration_test_coll"

// Helper to get a clean task collection for each test
func getTaskTestCollection(t *testing.T) *mongo.Collection {
	require.NotNil(t, testDBClient, "Database client not initialized. TestMain setup might have failed.")
	collection := testDBClient.Database(TestDatabaseName).Collection(testTaskCollectionName)
	_, err := collection.DeleteMany(context.Background(), bson.M{})
	require.NoError(t, err, "Failed to clean task test collection")
	return collection
}

func TestTaskRepository_Integration(t *testing.T) {
	if testDBClient == nil {
		t.Fatal("testDBClient is nil. TestMain setup for DB connection likely failed or was skipped.")
	}

	taskRepo := repositories.NewTaskRepository(testDBClient, TestDatabaseName, testTaskCollectionName)
	require.NotNil(t, taskRepo, "NewTaskRepository returned nil")

	ctx := context.Background()
	defaultUser := "integ_test_user"

	t.Run("NewTask_Success", func(t *testing.T) {
		_ = getTaskTestCollection(t) // Clean

		taskToCreate := domain.Task{
			Title:       "Integ Test Task 1",
			Description: "Description for task 1",
			Status:      "Pending",
			CreatedBy:   defaultUser,
			DueDate:     time.Now().Add(48 * time.Hour).Truncate(time.Millisecond), // Truncate for easier comparison
		}

		insertResult, err := taskRepo.NewTask(ctx, taskToCreate)
		require.NoError(t, err)
		require.NotNil(t, insertResult)
		insertedID, ok := insertResult.InsertedID.(primitive.ObjectID)
		require.True(t, ok)

		// Verify in DB
		var foundTask domain.Task
		collection := testDBClient.Database(TestDatabaseName).Collection(testTaskCollectionName)
		err = collection.FindOne(ctx, bson.M{"_id": insertedID}).Decode(&foundTask)
		require.NoError(t, err)
		assert.Equal(t, taskToCreate.Title, foundTask.Title)
		assert.Equal(t, taskToCreate.Description, foundTask.Description)
		assert.Equal(t, taskToCreate.Status, foundTask.Status)
		assert.Equal(t, taskToCreate.CreatedBy, foundTask.CreatedBy)
		assert.True(t, taskToCreate.DueDate.Equal(foundTask.DueDate), "Due dates should match")
		assert.Equal(t, insertedID, foundTask.ID)
	})

	t.Run("GetTaskByID_Success", func(t *testing.T) {
		taskCollection := getTaskTestCollection(t) // Clean and get

		taskID := primitive.NewObjectID()
		taskToInsert := domain.Task{
			ID:        taskID,
			Title:     "Find Me Task",
			Status:    "InProgress",
			CreatedBy: defaultUser,
		}
		_, err := taskCollection.InsertOne(ctx, taskToInsert)
		require.NoError(t, err)

		foundTask, err := taskRepo.GetTaskByID(ctx, taskID.Hex())
		require.NoError(t, err)
		assert.Equal(t, taskToInsert.Title, foundTask.Title)
		assert.Equal(t, taskID, foundTask.ID)
	})

	t.Run("GetTaskByID_NotFound", func(t *testing.T) {
		_ = getTaskTestCollection(t) // Clean
		nonExistentID := primitive.NewObjectID()
		_, err := taskRepo.GetTaskByID(ctx, nonExistentID.Hex())
		require.Error(t, err)
		assert.EqualError(t, err, "task not found") // Error from your repository
	})

	t.Run("GetTaskByID_InvalidIDFormat", func(t *testing.T) {
		_ = getTaskTestCollection(t)
		_, err := taskRepo.GetTaskByID(ctx, "this-is-not-an-object-id")
		require.Error(t, err)
		assert.EqualError(t, err, "invalid task ID format")
	})

	t.Run("GetAllTask_Success", func(t *testing.T) {
		taskCollection := getTaskTestCollection(t) // Clean and get

		task1 := domain.Task{ID: primitive.NewObjectID(), Title: "Task A", Status: "Todo", CreatedBy: defaultUser}
		task2 := domain.Task{ID: primitive.NewObjectID(), Title: "Task B", Status: "Done", CreatedBy: defaultUser}
		_, err := taskCollection.InsertMany(ctx, []interface{}{task1, task2})
		require.NoError(t, err)

		tasks, err := taskRepo.GetAllTask(ctx)
		require.NoError(t, err)
		require.Len(t, tasks, 2)
		// Check if both tasks are present (order might not be guaranteed by Find)
		assert.Contains(t, tasks, task1)
		assert.Contains(t, tasks, task2)
	})

	t.Run("GetAllTask_Empty", func(t *testing.T) {
		_ = getTaskTestCollection(t) // Clean
		tasks, err := taskRepo.GetAllTask(ctx)
		require.NoError(t, err)
		assert.Empty(t, tasks) // Or assert.Len(t, tasks, 0)
	})

	t.Run("UpdateTask_Success", func(t *testing.T) {
		taskCollection := getTaskTestCollection(t) // Clean and get

		originalTask := domain.Task{
			ID:        primitive.NewObjectID(),
			Title:     "Original Title",
			Status:    "Pending",
			CreatedBy: defaultUser,
		}
		_, err := taskCollection.InsertOne(ctx, originalTask)
		require.NoError(t, err)

		updateData := domain.Task{
			Title:       "Updated Integ Title",
			Description: "Updated Description",
			Status:      "Completed",
			DueDate:     time.Now().Add(72 * time.Hour).Truncate(time.Millisecond),
		}

		err = taskRepo.UpdateTask(ctx, originalTask.ID.Hex(), updateData)
		require.NoError(t, err)

		// Verify in DB
		var updatedTaskInDB domain.Task
		err = taskCollection.FindOne(ctx, bson.M{"_id": originalTask.ID}).Decode(&updatedTaskInDB)
		require.NoError(t, err)
		assert.Equal(t, updateData.Title, updatedTaskInDB.Title)
		assert.Equal(t, updateData.Description, updatedTaskInDB.Description)
		assert.Equal(t, updateData.Status, updatedTaskInDB.Status)
		assert.True(t, updateData.DueDate.Equal(updatedTaskInDB.DueDate))
		assert.Equal(t, originalTask.CreatedBy, updatedTaskInDB.CreatedBy) // CreatedBy should not change
	})

	t.Run("UpdateTask_TaskNotFound", func(t *testing.T) {
		_ = getTaskTestCollection(t) // Clean
		nonExistentID := primitive.NewObjectID()
		updateData := domain.Task{Title: "Won't Update"}
		err := taskRepo.UpdateTask(ctx, nonExistentID.Hex(), updateData)
		require.Error(t, err)
		assert.EqualError(t, err, "task not found")
	})

	t.Run("UpdateTask_NoFieldsToUpdate", func(t *testing.T) {
		taskCollection := getTaskTestCollection(t) // Clean and get
		taskToUpdate := domain.Task{ID: primitive.NewObjectID(), Title: "A Task"}
		_, err := taskCollection.InsertOne(ctx, taskToUpdate)
		require.NoError(t, err)

		emptyUpdate := domain.Task{} // No fields set
		err = taskRepo.UpdateTask(ctx, taskToUpdate.ID.Hex(), emptyUpdate)
		require.Error(t, err)
		assert.EqualError(t, err, "no field provided")
	})

	t.Run("DeleteTask_Success", func(t *testing.T) {
		taskCollection := getTaskTestCollection(t) // Clean and get

		taskToDelete := domain.Task{ID: primitive.NewObjectID(), Title: "To Be Deleted"}
		_, err := taskCollection.InsertOne(ctx, taskToDelete)
		require.NoError(t, err)

		err = taskRepo.DeleteTask(ctx, taskToDelete.ID.Hex())
		require.NoError(t, err)

		// Verify in DB
		count, err := taskCollection.CountDocuments(ctx, bson.M{"_id": taskToDelete.ID})
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "Task should be deleted from DB")
	})

	t.Run("DeleteTask_TaskNotFound", func(t *testing.T) {
		_ = getTaskTestCollection(t) // Clean
		nonExistentID := primitive.NewObjectID()
		err := taskRepo.DeleteTask(ctx, nonExistentID.Hex())
		require.Error(t, err)
		assert.EqualError(t, err, "task not found")
	})
}
