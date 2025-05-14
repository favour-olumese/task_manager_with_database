package infrastructure_test

import (
	"errors"
	"strings"
	infrastructure "task_manager/Infrastructure"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// Tests for PasswordService
func TestPasswordService(t *testing.T) {
	// Create an instance of our paassword service
	passwordService := infrastructure.NewPasswordService()
	require.NotNil(t, passwordService, "NewPasswordService should not return nil")

	t.Run("HashPassword_Success", func(t *testing.T) {
		password := "mySecurePassword123"
		hashedPassword, err := passwordService.HashPassword(password)

		// Assert that no error occured during hashing
		require.NoError(t, err, "HashPassword should not return an error on success")

		// Assert that the hashed passwprd is not empty
		require.NotEmpty(t, hashedPassword, "Hashed password should not be empty")

		// Assert that the hashed password is not the sae as the origin
		assert.NotEqual(t, password, hashedPassword, "Hashed password should not be the original password")

		// Check that it looks like bcrypt hash
		assert.True(t, strings.HasPrefix(hashedPassword, "$2a$") ||
			strings.HasPrefix(hashedPassword, "$2b$") ||
			strings.HasPrefix(hashedPassword, "$2y$"),
			"Hashed password should have a bcrypt prefix, got: %s", hashedPassword)

		assert.Greater(t, len(hashedPassword), 30, "bcrypt hash should be reasonably long")
	})

	t.Run("HashPassword_ErrorOnExtremelyLongPassword", func(t *testing.T) {
		// bcrypt has a maximum password length of 72 bytes.
		longPassword := strings.Repeat("a", 73)
		hashedPassword, err := passwordService.HashPassword(longPassword)

		require.Error(t, err, "HashPassword shouls return an error for passwords exceeding bcrypt's limit")
		assert.Empty(t, hashedPassword, "Hashed password should be empty on error")

		assert.EqualError(t, err, "failed to hash password", "Error message should match the wrapped error")

	})

	t.Run("ComparePasswords_Success_Matching", func(t *testing.T) {
		password := "mySecurePassword123"
		hashedPassword, err := passwordService.HashPassword(password)

		require.NoError(t, err, "Pre-condition: HashPassword failed")

		err = passwordService.ComparePasswords(hashedPassword, password)

		// Assert that no error occurred
		assert.NoError(t, err, "ComparePasswords should not return an error for matchinf passwords")

	})

	t.Run("ComparePasswords_Failure_NonMatching", func(t *testing.T) {
		password := "mySecurePassword123"
		wrongPassword := "ThisIsTheWrongPassword"

		// Hash the correct password
		hashedPassword, err := passwordService.HashPassword(password)
		require.NoError(t, err, "Pre-condition: HashPassword failed")

		err = passwordService.ComparePasswords(hashedPassword, wrongPassword)

		// Assert that an error occurred
		require.Error(t, err, "ComparePasswords should return an error for non-matching passwords")

		// Assert that the specific error is bcrypt's mismatch
		assert.ErrorIs(t, err, bcrypt.ErrMismatchedHashAndPassword, "Error should be bycrypt.ErrMismatchedHashAndPassword")

	})

	t.Run("ComparePasswords_Failure_InvalidHashFormat", func(t *testing.T) {
		password := "mySecurePassword123"
		invalidHash := "thisisnotavalidbcrypthashatall"

		// Attempt to compare with an invalidHash format
		err := passwordService.ComparePasswords(invalidHash, password)

		assert.Error(t, err, "ComparePasswords should return an error for an invalid hash format")

		// ComparePasswords directly returns the error from bcrypt.CompareHashAndPassword.
		if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) &&
			!errors.Is(err, bcrypt.ErrHashTooShort) {
			t.Errorf("Expected a bcrypt specific error or MismatchedHashAndPassword, but got: %v", err)
		}

		assert.ErrorIs(t, err, bcrypt.ErrHashTooShort, "For a short malformed hash, expect ErrHashTooShort")

	})

	t.Run("HashAndCompare_EmptyPassword", func(t *testing.T) {
		// bcrypt can hash an empty string.
		password := ""
		hashedPassword, err := passwordService.HashPassword(password)

		require.NoError(t, err, "HashPassword should handle empty passwords without error (by hashing the empty string)")
		require.NotEmpty(t, hashedPassword)

		// Verify that the hash of an empty string can be compared correctly
		err = passwordService.ComparePasswords(hashedPassword, password)
		assert.NoError(t, err, "Comparing hash of empty string with empty string should succeed")

		// Verify that comparing with a non-empty string fails
		err = passwordService.ComparePasswords(hashedPassword, "nonemptypass")
		assert.ErrorIs(t, err, bcrypt.ErrMismatchedHashAndPassword, "Comparing hash of empty string with non-empty string should fail")

	})

}
