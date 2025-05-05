package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
type ReisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}
