package usecases

import (
	"context"
	"errors"
	domain "task_manager/Domain"

	"go.mongodb.org/mongo-driver/mongo"
)

type userUsecase struct {
	userRepo        domain.UserRepository
	passwordService domain.PasswordService
	jwtService      domain.JWTService
}

func NewUserUsecase(repo domain.UserRepository, passwordService domain.PasswordService, jwtService domain.JWTService) domain.UserUsecase {
	return &userUsecase{
		userRepo:        repo,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

func (usecase *userUsecase) Register(ctx context.Context, username, password string) (*mongo.InsertOneResult, error) {
	_, err := usecase.userRepo.FindUserByUsername(ctx, username)

	// nil is returned if user already exist, else an error is returned.
	if err == nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := usecase.passwordService.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := domain.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Role:         domain.RoleUser,
	}

	// Save tp the database
	return usecase.userRepo.CreateUser(ctx, &user)
}

func (usecase *userUsecase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := usecase.userRepo.FindUserByUsername(ctx, username)

	// Find user
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	// Compare password
	if err := usecase.passwordService.ComparePasswords(user.PasswordHash, password); err != nil {
		return "", errors.New("invalid username or password")
	}

	// Generate JWT token
	token, err := usecase.jwtService.GenerateToken(username, user.Role)
	if err != nil {
		return "", err
	}
	return token, nil
}
