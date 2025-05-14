package usecases_test

import (
	"context"
	"errors"
	domain "task_manager/Domain"
	usecases "task_manager/Usecases"
	"task_manager/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Define the suite struct
type TaskUsecaseSuite struct {
	suite.Suite
	mockTaskRepo *mocks.MockTaskRepository
	taskUsecase  domain.TaskUsecase
}

// Setup runs before each test in the suite
func (s *TaskUsecaseSuite) SetupTest() {
	s.mockTaskRepo = mocks.NewMockTaskRepository(s.T())
	s.taskUsecase = usecases.NewTaskUsecase(s.mockTaskRepo)
}

// Runs the entire suite
func TestTaskUsecaseSuite(t *testing.T) {
	suite.Run(t, new(TaskUsecaseSuite))
}

// ---- Test GetAllTask ----

func (s *TaskUsecaseSuite) TestGetAllTask_Success() {
	ctx := context.Background()
	expectedTasks := []domain.Task{
		{ID: primitive.NewObjectID(), Title: "Task 1", Status: "Pending"},
		{ID: primitive.NewObjectID(), Title: "Task 2", Status: "Done"},
	}

	// Arrange
	s.mockTaskRepo.EXPECT().
		GetAllTask(ctx).
		Return(expectedTasks, nil).
		Once()

	// Act
	tasks, err := s.taskUsecase.GetAllTask(ctx)

	// Assert
	s.NoError(err)
	s.Equal(expectedTasks, tasks)

}

func (s *TaskUsecaseSuite) TestGetAllTask_RepositoryError() {
	ctx := context.Background()
	repoError := errors.New("database error")

	// Arrange
	s.mockTaskRepo.EXPECT().
		GetAllTask(ctx).
		Return(nil, repoError).
		Once()

	// Act
	tasks, err := s.taskUsecase.GetAllTask(ctx)

	// Assert
	s.Error(err)
	s.Nil(tasks)
	s.Equal(repoError, err)

}

func (s *TaskUsecaseSuite) TestGetAllTask_NoTasks() {
	ctx := context.Background()
	expectedTasks := []domain.Task{} // Empty slice

	// Arrange
	s.mockTaskRepo.EXPECT().
		GetAllTask(ctx).
		Return(expectedTasks, nil).
		Once()

	// Act
	tasks, err := s.taskUsecase.GetAllTask(ctx)

	// Assert
	s.NoError(err)
	s.NotNil(tasks) // Should return an empty slice, not nil
	s.Len(tasks, 0) // Verify length is 0
	s.Equal(expectedTasks, tasks)

}

// ---- Test GetTaskByID ----

func (s *TaskUsecaseSuite) TestGetTaskByID_Success() {
	ctx := context.Background()
	taskID := primitive.NewObjectID()
	expectedTasks := domain.Task{ID: taskID, Title: "Specific Task", Status: "In Progress"}

	// Arrange
	s.mockTaskRepo.EXPECT().
		GetTaskByID(ctx, taskID.Hex()).
		Return(expectedTasks, nil).
		Once()

	// Act
	task, err := s.taskUsecase.GetTaskByID(ctx, taskID.Hex())

	// Assert
	s.NoError(err)
	s.Equal(expectedTasks, task)

}

func (s *TaskUsecaseSuite) TestGetTaskByID_NotFound() {
	ctx := context.Background()
	taskID := primitive.NewObjectID()
	notFoundError := errors.New("task not found") // Error returned by repo

	// Arrange
	s.mockTaskRepo.EXPECT().
		GetTaskByID(ctx, taskID.Hex()).
		Return(domain.Task{}, notFoundError).
		Once()

	// Act
	task, err := s.taskUsecase.GetTaskByID(ctx, taskID.Hex())

	// Assert
	s.Error(err)
	s.Equal(notFoundError, err)
	s.Equal(domain.Task{}, task) // Expect zero value task

}

func (s *TaskUsecaseSuite) TestGetTaskByID_RepositoryError() {
	ctx := context.Background()
	taskID := primitive.NewObjectID()
	repoError := errors.New("internal server error")

	// Arrange
	s.mockTaskRepo.EXPECT().
		GetTaskByID(ctx, taskID.Hex()).
		Return(domain.Task{}, repoError).
		Once()

	// Act
	task, err := s.taskUsecase.GetTaskByID(ctx, taskID.Hex())

	// Assert
	s.Error(err)
	s.Equal(repoError, err)
	s.Equal(domain.Task{}, task)

}

// ---- Test UpdateTask ----

func (s *TaskUsecaseSuite) TestUpdateTask_Success() {
	ctx := context.Background()
	taskID := primitive.NewObjectID()
	updatedTask := domain.Task{Title: "Updated Title", Status: "Done"}

	// Arrange
	// The mock expects the full updatedTask struct as passed from the usecase
	s.mockTaskRepo.EXPECT().
		UpdateTask(ctx, taskID.Hex(), updatedTask).
		Return(nil).
		Once()

	// Act
	err := s.taskUsecase.UpdateTask(ctx, taskID.Hex(), updatedTask)

	// Assert
	s.NoError(err)

}

func (s *TaskUsecaseSuite) TestUpdateTask_RepositoryError() {
	ctx := context.Background()
	taskID := primitive.NewObjectID()
	updatedTask := domain.Task{Title: "Updated Title", Status: "Done"}
	repoError := errors.New("update failed")

	// Arrange
	s.mockTaskRepo.EXPECT().
		UpdateTask(ctx, taskID.Hex(), updatedTask).
		Return(repoError).
		Once()

	// Act
	err := s.taskUsecase.UpdateTask(ctx, taskID.Hex(), updatedTask)

	// Assert
	s.Error(err)
	s.Equal(repoError, err)

}

// ---- Test DeleteTask ----
// This can be improved when authoriaztion is made a requirement for deletion of tasks

func (s *TaskUsecaseSuite) TestDeleteTask_Success() {
	ctx := context.Background()
	taskID := primitive.NewObjectID()

	// Arrange
	s.mockTaskRepo.EXPECT().
		DeleteTask(ctx, taskID.Hex()).
		Return(nil).
		Once()

	// Act
	err := s.taskUsecase.DeleteTask(ctx, taskID.Hex())

	// Assert
	s.NoError(err)

}

func (s *TaskUsecaseSuite) TestDeleteTask_RepositoryError() {
	ctx := context.Background()
	taskID := primitive.NewObjectID()
	repoError := errors.New("deletion failed")

	// Arrange
	s.mockTaskRepo.EXPECT().
		DeleteTask(ctx, taskID.Hex()).
		Return(repoError).
		Once()

	// Act
	err := s.taskUsecase.DeleteTask(ctx, taskID.Hex())

	// Assert
	s.Error(err)
	s.Equal(repoError, err)

}

// ---- Test NewTask ----

func (s *TaskUsecaseSuite) TestNewTask_Success() {
	ctx := context.Background()
	newTask := domain.Task{
		Title:       "New Task",
		Description: "Details",
		Status:      "Todo",
		DueDate:     time.Now().Add(24 * time.Hour),
		CreatedBy:   "testuser", // Assuming this is set before calling usecase
	}

	mockObjectID := primitive.NewObjectID()
	mockInsertResult := &mongo.InsertOneResult{InsertedID: mockObjectID}

	// Arrange
	// The exact newTask struct is to be passed to the repository
	s.mockTaskRepo.EXPECT().
		NewTask(ctx, newTask).
		Return(mockInsertResult, nil).
		Once()

	// Act
	result, err := s.taskUsecase.NewTask(ctx, newTask)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal(mockObjectID, result.InsertedID)

}

func (s *TaskUsecaseSuite) TestNewTask_RepositoryError() {
	ctx := context.Background()
	newTask := domain.Task{Title: "Failing Task", Status: "Todo"}
	repoError := errors.New("failed to insert task")

	// Arrange
	s.mockTaskRepo.EXPECT().
		NewTask(ctx, newTask).
		Return(nil, repoError).
		Once()

	// Act
	result, err := s.taskUsecase.NewTask(ctx, newTask)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Equal(repoError, err)

}
