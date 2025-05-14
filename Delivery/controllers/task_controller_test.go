package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"task_manager/Delivery/controllers"
	domain "task_manager/Domain" // For GetUserFromContext simulation
	"task_manager/mocks"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Helper to set up a Gin router with the TaskController for testing
func setupTaskRouter(taskUsecase domain.TaskUsecase) (*gin.Engine, *controllers.TaskController) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	taskController := controllers.NewTaskController(taskUsecase)
	return router, taskController
}

// Helper to add authenticated user to Gin context for testing protected routes
func addAuthToContext(c *gin.Context, username string, role string) {
	c.Set("username", username)
	c.Set("role", role)
}

// --- Test TaskController ---
func TestTaskController_GetAllTask(t *testing.T) {
	mockUsecase := mocks.NewMockTaskUsecase(t) // Adjust constructor if needed
	router, taskController := setupTaskRouter(mockUsecase)
	router.GET("/tasks", func(c *gin.Context) { // Simulate Auth middleware
		addAuthToContext(c, "testuser", domain.RoleUser)
		taskController.GetAllTask(c)
	})

	t.Run("Success_ReturnsTasks", func(t *testing.T) {
		// Arrange
		expectedTasks := []domain.Task{
			{ID: primitive.NewObjectID(), Title: "Task 1", CreatedBy: "testuser"},
			{ID: primitive.NewObjectID(), Title: "Task 2", CreatedBy: "testuser"},
		}
		mockUsecase.EXPECT().
			GetAllTask(mock.Anything). // Controller passes c.Request.Context()
			Return(expectedTasks, nil).
			Once()

		req, _ := http.NewRequest(http.MethodGet, "/tasks", nil)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		var tasks []domain.Task
		err := json.Unmarshal(rr.Body.Bytes(), &tasks)
		require.NoError(t, err)
		assert.Equal(t, expectedTasks, tasks)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("InternalServerError_UsecaseError", func(t *testing.T) {
		// Arrange
		usecaseError := errors.New("db query failed")
		mockUsecase.EXPECT().
			GetAllTask(mock.Anything).
			Return(nil, usecaseError).
			Once()

		req, _ := http.NewRequest(http.MethodGet, "/tasks", nil)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, usecaseError.Error(), respBody["error"])
		mockUsecase.AssertExpectations(t)
	})
}

func TestTaskController_GetTaskByID(t *testing.T) {
	mockUsecase := mocks.NewMockTaskUsecase(t)
	router, taskController := setupTaskRouter(mockUsecase)
	// Path parameter :id
	router.GET("/tasks/:id", func(c *gin.Context) {
		addAuthToContext(c, "testuser", domain.RoleUser)
		taskController.GetTaskByID(c)
	})

	t.Run("Success_ReturnsTask", func(t *testing.T) {
		// Arrange
		taskID := primitive.NewObjectID()
		expectedTask := domain.Task{ID: taskID, Title: "Specific Task", CreatedBy: "testuser"}

		mockUsecase.EXPECT().
			GetTaskByID(mock.Anything, taskID.Hex()).
			Return(expectedTask, nil).
			Once()

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/tasks/%s", taskID.Hex()), nil)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		var task domain.Task
		json.Unmarshal(rr.Body.Bytes(), &task)
		assert.Equal(t, expectedTask, task)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("NotFound_UsecaseReturnsError", func(t *testing.T) {
		// Arrange
		taskID := primitive.NewObjectID()
		// Your GetTaskByID controller maps usecase errors to http.StatusNotFound
		usecaseError := errors.New("task not found in db") // Usecase returns this

		mockUsecase.EXPECT().
			GetTaskByID(mock.Anything, taskID.Hex()).
			Return(domain.Task{}, usecaseError). // Return empty task and error
			Once()

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/tasks/%s", taskID.Hex()), nil)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, usecaseError.Error(), respBody["error"])
		mockUsecase.AssertExpectations(t)
	})
}

func TestTaskController_NewTask(t *testing.T) {
	mockUsecase := mocks.NewMockTaskUsecase(t)
	router, taskController := setupTaskRouter(mockUsecase)
	router.POST("/tasks", func(c *gin.Context) {
		addAuthToContext(c, "newtaskuser", domain.RoleUser) // User creating the task
		taskController.NewTask(c)
	})

	t.Run("Success_CreatesTask", func(t *testing.T) {
		// Arrange
		newTaskReq := domain.Task{Title: "A New Task", Description: "Description here", Status: "Pending"}
		reqBodyBytes, _ := json.Marshal(newTaskReq)

		mockObjectID := primitive.NewObjectID()
		// Your controller expects *mongo.InsertOneResult from usecase.NewTask
		mockInsertResult := &mongo.InsertOneResult{InsertedID: mockObjectID}

		// The controller will set CreatedBy from context, then call usecase
		expectedTaskToUsecase := newTaskReq
		expectedTaskToUsecase.CreatedBy = "newtaskuser" // This is what controller passes

		mockUsecase.EXPECT().
			NewTask(mock.Anything, expectedTaskToUsecase).
			Return(mockInsertResult, nil).
			Once()

		req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
		var createdTask domain.Task
		json.Unmarshal(rr.Body.Bytes(), &createdTask)

		// The controller sets the ID on the newTask object after successful insertion
		assert.Equal(t, mockObjectID, createdTask.ID)
		assert.Equal(t, newTaskReq.Title, createdTask.Title)
		assert.Equal(t, "newtaskuser", createdTask.CreatedBy) // Check if CreatedBy is correctly set in response
		mockUsecase.AssertExpectations(t)
	})

	t.Run("BadRequest_InvalidJSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req) // Will hit the auth simulation first
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockUsecase.AssertNotCalled(t, "NewTask")
	})

	t.Run("InternalServerError_GetUserFromContextError", func(t *testing.T) {
		// Arrange: Simulate GetUserFromContext failing by not setting context values
		// We need a new router instance for this specific handler setup
		freshMockUsecase := mocks.NewMockTaskUsecase(t)
		freshRouter, freshTaskController := setupTaskRouter(freshMockUsecase)
		freshRouter.POST("/tasks_no_auth_ctx", freshTaskController.NewTask) // No addAuthToContext call

		newTaskReq := domain.Task{Title: "Task Title"}
		reqBodyBytes, _ := json.Marshal(newTaskReq)
		req, _ := http.NewRequest(http.MethodPost, "/tasks_no_auth_ctx", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		freshRouter.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		// Your GetUserFromContext returns "username is not found in context"
		assert.Equal(t, "username is not found in context", respBody["error"])
		freshMockUsecase.AssertNotCalled(t, "NewTask", mock.Anything, mock.Anything)
	})

	t.Run("InternalServerError_UsecaseError", func(t *testing.T) {
		// Arrange
		newTaskReq := domain.Task{Title: "A New Task"}
		reqBodyBytes, _ := json.Marshal(newTaskReq)
		usecaseError := errors.New("failed to insert task in db")

		expectedTaskToUsecase := newTaskReq
		expectedTaskToUsecase.CreatedBy = "newtaskuser"

		mockUsecase.EXPECT().
			NewTask(mock.Anything, expectedTaskToUsecase).
			Return(nil, usecaseError).
			Once()

		req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, usecaseError.Error(), respBody["error"])
		mockUsecase.AssertExpectations(t)
	})
}

func TestTaskController_UpdateTask(t *testing.T) {
	mockUsecase := mocks.NewMockTaskUsecase(t)
	router, taskController := setupTaskRouter(mockUsecase)
	router.PUT("/tasks/:id", func(c *gin.Context) {
		addAuthToContext(c, "taskupdater", domain.RoleUser)
		taskController.UpdateTask(c)
	})

	t.Run("Success_UpdatesTask", func(t *testing.T) {
		// Arrange
		taskID := primitive.NewObjectID()
		updateReq := domain.Task{Title: "Updated Title", Status: "Completed"} // Only send fields to update
		reqBodyBytes, _ := json.Marshal(updateReq)

		mockUsecase.EXPECT().
			UpdateTask(mock.Anything, taskID.Hex(), updateReq).
			Return(nil). // Successful update returns nil error
			Once()

		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/tasks/%s", taskID.Hex()), bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, "Task updated", respBody["message"])
		mockUsecase.AssertExpectations(t)
	})

	t.Run("BadRequest_InvalidJSON", func(t *testing.T) {
		taskID := primitive.NewObjectID()
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/tasks/%s", taskID.Hex()), bytes.NewBufferString(`{`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockUsecase.AssertNotCalled(t, "UpdateTask")
	})

	t.Run("InternalServerError_UsecaseError", func(t *testing.T) {
		// Arrange
		taskID := primitive.NewObjectID()
		updateReq := domain.Task{Title: "Update Title"}
		reqBodyBytes, _ := json.Marshal(updateReq)
		usecaseError := errors.New("update failed in db")

		mockUsecase.EXPECT().
			UpdateTask(mock.Anything, taskID.Hex(), updateReq).
			Return(usecaseError).
			Once()

		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/tasks/%s", taskID.Hex()), bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, usecaseError.Error(), respBody["error"])
		mockUsecase.AssertExpectations(t)
	})
}

func TestTaskController_DeleteTask(t *testing.T) {
	mockUsecase := mocks.NewMockTaskUsecase(t)
	router, taskController := setupTaskRouter(mockUsecase)
	router.DELETE("/tasks/:id", func(c *gin.Context) {
		addAuthToContext(c, "taskdeleter", domain.RoleUser)
		taskController.DeleteTask(c)
	})

	t.Run("Success_DeletesTask", func(t *testing.T) {
		// Arrange
		taskID := primitive.NewObjectID()
		mockUsecase.EXPECT().
			DeleteTask(mock.Anything, taskID.Hex()).
			Return(nil). // Successful delete returns nil error
			Once()

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/tasks/%s", taskID.Hex()), nil)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, "Task deleted successfully", respBody["message"])
		mockUsecase.AssertExpectations(t)
	})

	t.Run("InternalServerError_UsecaseError", func(t *testing.T) {
		// Arrange
		taskID := primitive.NewObjectID()
		usecaseError := errors.New("delete failed in db")

		mockUsecase.EXPECT().
			DeleteTask(mock.Anything, taskID.Hex()).
			Return(usecaseError).
			Once()

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/tasks/%s", taskID.Hex()), nil)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, usecaseError.Error(), respBody["error"])
		mockUsecase.AssertExpectations(t)
	})
}
