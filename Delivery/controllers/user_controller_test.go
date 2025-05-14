package controllers_test // Create a _test package

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"task_manager/Delivery/controllers"
	domain "task_manager/Domain"
	"task_manager/mocks"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Helper to set up a Gin router with the UserController for testing
func setupUserRouter(userUsecase domain.UserUsecase) (*gin.Engine, *controllers.UserController) {
	gin.SetMode(gin.TestMode) // Set Gin to test mode
	router := gin.New()       // Use gin.New() for a clean router, not gin.Default() which includes middleware
	userController := controllers.NewUserController(userUsecase)
	return router, userController
}

// --- Test UserController ---
func TestUserController_Register(t *testing.T) {
	// --- Setup ---
	// Create a mock UserUsecase. Adjust NewMockUserUsecase if your generated mock has a different constructor name.
	mockUsecase := mocks.NewMockUserUsecase(t) // Assuming this is the constructor name
	router, userController := setupUserRouter(mockUsecase)

	// Define the route for the handler we are testing
	router.POST("/users/register", userController.Register)

	// --- Test Cases ---
	t.Run("Success", func(t *testing.T) {
		// Arrange
		registerReq := domain.RegisterRequest{
			Username: "newtestuser",
			Password: "password123",
		}
		reqBodyBytes, _ := json.Marshal(registerReq)

		mockObjectID := primitive.NewObjectID()
		// Expects a *mongo.InsertOneResult from the usecase's Register method
		mockInsertResult := &mongo.InsertOneResult{InsertedID: mockObjectID}

		// Expect the Register method on the mock usecase to be called
		mockUsecase.EXPECT().
			Register(mock.AnythingOfType("*context.timerCtx"), registerReq.Username, registerReq.Password).
			Return(mockInsertResult, nil). // Return success
			Once()

		// Create an HTTP request
		req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder() // Create a response recorder

		// Act: Serve the HTTP request to the router
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code, "Response code should be 201 Created")

		var responseBody map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
		require.NoError(t, err, "Response body should be valid JSON")

		assert.Equal(t, "User registered successfully", responseBody["message"])
		userData, ok := responseBody["user"].(map[string]interface{})
		require.True(t, ok, "'user' field should be a map")
		assert.Equal(t, mockObjectID.Hex(), userData["id"])
		assert.Equal(t, registerReq.Username, userData["username"])
		assert.Equal(t, domain.RoleUser, userData["role"]) // As per your controller's response

		mockUsecase.AssertExpectations(t) // Verify that all expected mock calls were made
	})

	t.Run("BadRequest_InvalidJSON", func(t *testing.T) {
		// Arrange: Send malformed JSON
		req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBufferString(`{"username": "test",`)) // Malformed
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)

	})

	t.Run("BadRequest_ValidationFailure", func(t *testing.T) {
		// Arrange: Send data that fails Gin's struct validation (e.g., password too short)
		registerReq := domain.RegisterRequest{Username: "validuser", Password: "123"} // Password "123" < 6 chars
		reqBodyBytes, _ := json.Marshal(registerReq)
		req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Contains(t, respBody["error"], "Password", "Error message should mention Password validation")

	})

	t.Run("InternalServerError_UsecaseError", func(t *testing.T) {
		// Arrange
		registerReq := domain.RegisterRequest{Username: "testuser", Password: "password123"}
		reqBodyBytes, _ := json.Marshal(registerReq)
		usecaseError := errors.New("simulated usecase db error")

		mockUsecase.EXPECT().
			Register(mock.AnythingOfType("*context.timerCtx"), registerReq.Username, registerReq.Password).
			Return(nil, usecaseError). // Usecase returns an error
			Once()

		req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		// Your controller currently maps all usecase errors to 500 in Register.
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, usecaseError.Error(), respBody["error"])
		mockUsecase.AssertExpectations(t)
	})

}

func TestUserController_Login(t *testing.T) {
	// --- Setup ---
	mockUsecase := mocks.NewMockUserUsecase(t)
	router, userController := setupUserRouter(mockUsecase)
	router.POST("/users/login", userController.Login)

	// --- Test Cases ---
	t.Run("Success", func(t *testing.T) {
		// Arrange
		loginReq := domain.LoginRequest{Username: "testuser", Password: "password123"}
		reqBodyBytes, _ := json.Marshal(loginReq)
		expectedToken := "valid.jwt.token.string"

		mockUsecase.EXPECT().
			Login(mock.AnythingOfType("*context.timerCtx"), loginReq.Username, loginReq.Password).
			Return(expectedToken, nil).
			Once()

		req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, expectedToken, respBody["token"])
		mockUsecase.AssertExpectations(t)
	})

	t.Run("BadRequest_InvalidJSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBufferString(`{"user":`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

	})

	t.Run("BadRequest_ValidationFailure_MissingFields", func(t *testing.T) {
		// Gin's 'required' binding tag
		loginReq := domain.LoginRequest{Username: "testuser"} // Missing password
		reqBodyBytes, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

	})

	t.Run("Unauthorized_UsecaseError_InvalidCredentials", func(t *testing.T) {
		// Arrange
		loginReq := domain.LoginRequest{Username: "wronguser", Password: "wrongpassword"}
		reqBodyBytes, _ := json.Marshal(loginReq)
		usecaseError := errors.New("invalid username or password") // Error from usecase

		mockUsecase.EXPECT().
			Login(mock.AnythingOfType("*context.timerCtx"), loginReq.Username, loginReq.Password).
			Return("", usecaseError).
			Once()

		req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		// Your Login controller correctly returns http.StatusUnauthorized for usecase errors.
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		var respBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, usecaseError.Error(), respBody["error"])
		mockUsecase.AssertExpectations(t)
	})

}
