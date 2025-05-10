package infrastructure

import (
	"errors"
	domain "task_manager/Domain"

	"golang.org/x/crypto/bcrypt"
)

type passwordService struct{}

func NewPasswordService() domain.PasswordService {
	return &passwordService{}
}

// Generate a hash of the password
func (service *passwordService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("failed to hash password")
	}
	return string(hashedPassword), nil
}

// Compare a hashed password with a plaintext password
func (service *passwordService) ComparePasswords(hashedPassword, plaintextPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plaintextPassword))

	return err
}
