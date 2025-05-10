package domain

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Task structure in the database.
type Task struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	DueDate     time.Time          `json:"due_date,omitempty" bson:"due_date,omitempty"`
	Status      string             `json:"status" bson:"status"`
	CreatedBy   string             `json:"created_by,omitempty" bson:"created_by,omitempty"`
}

// User in the database
type User struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username     string             `json:"username" bson:"username" binding:"required,min=3,max=50"`
	PasswordHash string             `json:"-" bson:"password_hash"` // "-" is used to exclude from JSON marshalling for security.
	Role         string             `json:"role" bson:"role"`       // e.g., "user, "admin
	// Phone        string             `json:"phone,omitempty" bson:"phone,omitempty"`
	// Email        string             `json:"email" bson:"email" binding:"required,email"`
}

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Using only during registraton
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ------------------------- Repository -------------------------
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (*mongo.InsertOneResult, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
}

// ------------------------- Infrastructure -------------------------

// JWTService Interface
type JWTService interface {
	GenerateToken(username, role string) (string, error)
	ValidateToken(token string) (*CustomClaims, error)
}

type CustomClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Password Service Interface
type PasswordService interface {
	HashPassword(password string) (string, error)
	ComparePasswords(hashedPassword, plaintextPassword string) error
}

// ------------------------- Usecase -------------------------

type UserUsecase interface {
	Register(ctx context.Context, username, password string) (*mongo.InsertOneResult, error)
	Login(ctx context.Context, username, password string) (string, error)
}

type TaskUsecase interface {
	GetAllTask(ctx context.Context) ([]Task, error)
	GetTaskByID(ctx context.Context, id string) (Task, error)
	UpdateTask(ctx context.Context, id string, updatedTask Task) error
	DeleteTask(ctx context.Context, id string) error
	NewTask(ctx context.Context, task Task) (*mongo.InsertOneResult, error)
}
