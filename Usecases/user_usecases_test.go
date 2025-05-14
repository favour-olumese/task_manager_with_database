package usecases_test

import (
	"context"
	"errors"
	domain "task_manager/Domain"
	usecases "task_manager/Usecases"
	"task_manager/mocks"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Suite struct
type UserUsecaseSuite struct {
	suite.Suite
	mockUserRepo        *mocks.MockUserRepository
	mockPasswordService *mocks.MockPasswordService
	mockJwtService      *mocks.MockJWTService
	userUsecase         domain.UserUsecase
}

// SetupTest run before each test in the suite
func (s *UserUsecaseSuite) SetupTest() {
	s.mockUserRepo = mocks.NewMockUserRepository(s.T())
	s.mockPasswordService = mocks.NewMockPasswordService(s.T())
	s.mockJwtService = mocks.NewMockJWTService(s.T())
	s.userUsecase = usecases.NewUserUsecase(s.mockUserRepo, s.mockPasswordService, s.mockJwtService)
}

// Runs the entire suite
func TestUserUsecaseSuite(t *testing.T) {
	suite.Run(t, new(UserUsecaseSuite))
}

// ---- Test Register ----

func (s *UserUsecaseSuite) TestRegister_Success() {
	ctx := context.Background()
	username := "testuser"
	password := "password123"
	hashedPassword := "hashed_password"
	mockObjectID := primitive.NewObjectID()
	mockInsertResult := &mongo.InsertOneResult{InsertedID: mockObjectID}

	// Arrange: setup mok expectations
	s.mockUserRepo.EXPECT().
		FindUserByUsername(ctx, username).
		Return(nil, mongo.ErrNoDocuments).
		Once() // Expect user not found initially

	s.mockPasswordService.EXPECT().
		HashPassword(password).
		Return(hashedPassword, nil) // Expect password hashing to succeed

	s.mockUserRepo.EXPECT().CreateUser(ctx, mock.MatchedBy(func(user *domain.User) bool {
		return user.Username == username &&
			user.PasswordHash == hashedPassword &&
			user.Role == domain.RoleUser
	})).
		Return(mockInsertResult, nil).
		Once() // Expect user creation to succeed.

	// Act: Call the method under test
	result, err := s.userUsecase.Register(ctx, username, password)

	// Assert: Verify the outcomes
	s.NoError(err)
	s.NotNil(result)
	s.Equal(mockObjectID, result.InsertedID)
	// AssertExpectations(s.T()) is automatically called by t.Cleanup
}

func (s *UserUsecaseSuite) TestRegister_UserAlreadyExists_OnFind() {
	ctx := context.Background()
	username := "existinguser"
	password := "password123"
	existingUser := &domain.User{Username: username}

	// Arrange
	s.mockUserRepo.EXPECT().
		FindUserByUsername(ctx, username).
		Return(existingUser, nil).
		Once()

	// Act
	result, err := s.userUsecase.Register(ctx, username, password)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.EqualError(err, "user already exists")
	s.mockPasswordService.AssertNotCalled(s.T(), "HashPassword", mock.Anything, mock.Anything) // Ensure hashing wasn't called
	s.mockUserRepo.AssertNotCalled(s.T(), "CreateUser", mock.Anything, mock.Anything)          // Ensure creation wasn't called

}

func (s *UserUsecaseSuite) TestRegister_PasswordHashingError() {
	ctx := context.Background()
	username := "testuser"
	password := "password123"
	hashError := errors.New("hashing failed")

	// Arrange
	s.mockUserRepo.EXPECT().
		FindUserByUsername(ctx, username).
		Return(nil, mongo.ErrNoDocuments).
		Once()

	s.mockPasswordService.EXPECT().
		HashPassword(password).
		Return("", hashError).
		Once() // Expect hashing to fail

	// Act
	result, err := s.userUsecase.Register(ctx, username, password)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Equal(hashError, err)
	s.mockUserRepo.AssertNotCalled(s.T(), "CreateUser", mock.Anything, mock.Anything)

}

func (s *UserUsecaseSuite) TestRegister_CreateUserError() {
	ctx := context.Background()
	username := "testuser"
	password := "password123"
	hashedPassword := "hashed_password"
	dbError := errors.New("databased connection error")

	// Arrange
	s.mockUserRepo.EXPECT().
		FindUserByUsername(ctx, username).
		Return(nil, mongo.ErrNoDocuments).
		Once()

	s.mockPasswordService.EXPECT().
		HashPassword(password).
		Return(hashedPassword, nil).
		Once()

	s.mockUserRepo.EXPECT().
		CreateUser(ctx, mock.MatchedBy(func(user *domain.User) bool {
			return user.Username == username
		})).
		Return(nil, dbError). // Expect user creation to fail
		Once()

	// Act
	result, err := s.userUsecase.Register(ctx, username, password)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Equal(dbError, err)
}

// ---- Test Login ----

func (s *UserUsecaseSuite) TestLogin_Success() {
	ctx := context.Background()
	username := "testuser"
	password := "password123"
	hashedPassword := "hashed_password" // Assume this matches the password
	role := domain.RoleUser
	expectedToken := "valid.jwt.token"
	foundUser := &domain.User{
		ID:           primitive.NewObjectID(),
		Username:     username,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	// Arrange
	s.mockUserRepo.EXPECT().
		FindUserByUsername(ctx, username).
		Return(foundUser, nil).
		Once()

	s.mockPasswordService.EXPECT().
		ComparePasswords(hashedPassword, password).
		Return(nil). // Expect passwords to match
		Once()

	s.mockJwtService.EXPECT().
		GenerateToken(username, role).
		Return(expectedToken, nil). // Expect token generation to succeed
		Once()

	// Act
	token, err := s.userUsecase.Login(ctx, username, password)

	// Assert
	s.NoError(err)
	s.Equal(expectedToken, token)

}

func (s *UserUsecaseSuite) TestLogin_UserNotFound() {
	ctx := context.Background()
	username := "nonexistentuser"
	password := "password123"

	// Arrange
	s.mockUserRepo.EXPECT().
		FindUserByUsername(ctx, username).
		Return(nil, mongo.ErrNoDocuments). // Expect user not found
		Once()

	// Act
	token, err := s.userUsecase.Login(ctx, username, password)

	// Assert
	s.Error(err)
	s.Empty(token)
	s.EqualError(err, "invalid username or password")
	s.mockPasswordService.AssertNotCalled(s.T(), "ComparePasswords", mock.Anything, mock.Anything)
	s.mockJwtService.AssertNotCalled(s.T(), "GenerateToken", mock.Anything, mock.Anything)

}

func (s *UserUsecaseSuite) TestLogin_IncorrectPassword() {
	ctx := context.Background()
	username := "testuser"
	correctPasswordHash := "correct_hash"
	incorrectPassword := "wrongpassword"
	role := domain.RoleUser
	foundUser := &domain.User{
		ID:           primitive.NewObjectID(),
		Username:     username,
		PasswordHash: correctPasswordHash,
		Role:         role,
	}

	// Arrange
	s.mockUserRepo.EXPECT().
		FindUserByUsername(ctx, username).
		Return(foundUser, nil).
		Once()

	s.mockPasswordService.EXPECT().
		ComparePasswords(correctPasswordHash, incorrectPassword).
		Return(bcrypt.ErrMismatchedHashAndPassword). // Expect password comparison to fail.
		Once()

	// Act
	token, err := s.userUsecase.Login(ctx, username, incorrectPassword)

	// Assert
	s.Error(err)
	s.Empty(token)
	s.EqualError(err, "invalid username or password")
	s.mockJwtService.AssertNotCalled(s.T(), "GenerateToken", mock.Anything, mock.Anything)

}

func (s *UserUsecaseSuite) TestLogin_TokenGenerationError() {
	ctx := context.Background()
	username := "testuser"
	password := "password123"
	hashedPassword := "hashed_password"
	role := domain.RoleUser
	tokenError := errors.New("failed to sign token")
	foundUser := &domain.User{
		ID:           primitive.NewObjectID(),
		Username:     username,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	// Arrange
	s.mockUserRepo.EXPECT().
		FindUserByUsername(ctx, username).
		Return(foundUser, nil).
		Once()

	s.mockPasswordService.EXPECT().
		ComparePasswords(hashedPassword, password).
		Return(nil).
		Once()

	s.mockJwtService.EXPECT().
		GenerateToken(username, role).
		Return("", tokenError).
		Once()

	// Act
	token, err := s.userUsecase.Login(ctx, username, password)

	// Assert
	s.Error(err)
	s.Empty(token)
	s.Equal(tokenError, err)

}
