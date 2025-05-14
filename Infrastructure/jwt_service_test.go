package infrastructure_test

import (
	domain "task_manager/Domain"
	infrastructure "task_manager/Infrastructure"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTServie(t *testing.T) {
	jwtService := infrastructure.NewJWTService()

	require.NotNil(t, jwtService, "NewJWTService should not retuen nil")

	username := "testuser"
	role := domain.RoleUser

	// ---- Test GenerateToken ----
	t.Run("GenerateToken_Success", func(t *testing.T) {
		tokenString, err := jwtService.GenerateToken(username, role)

		require.NoError(t, err, "GenerateToken should not return an error on success")
		require.NotEmpty(t, tokenString, "Generated token string should not be empty")

		// Parse the token
		parsedToken, parseErr := jwt.ParseWithClaims(tokenString, &domain.CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte("ahnljdbjiohwebljnsknpihdbuo"), nil
		})

		require.NoError(t, parseErr, "Generated token should be parsable with the correct secret")
		require.NotNil(t, parsedToken, "Parsed token should not be nil")
		require.True(t, parsedToken.Valid, "Generated token should be valid immediately after generation")

		claims, ok := parsedToken.Claims.(*domain.CustomClaims)
		require.True(t, ok, "Token claims should be of type *domain.CustomClaims")

		assert.Equal(t, username, claims.Username, "Username in claims should match")
		assert.Equal(t, role, claims.Role, "Role in claims should match")
		assert.Equal(t, username, claims.Subject, "Subject in claims should match username")

		// Check timestamps (allowinng for a small delta due to processing time)
		// Service generates tokens valid for 24 hours
		expectedExpiry := time.Now().Add(24 * time.Hour)
		assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 5*time.Second, "Expiration time should be 24 hours from now")
		assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, 5*time.Second, "IssuedAT time should be around now")

	})

	// t.Run("GenerateToken_ErrorFromSigning", func(t *testing.T) {})

	t.Run("ValidateToken_Success_ValidToken", func(t *testing.T) {
		// Generate a token
		validTokenString, genErr := jwtService.GenerateToken(username, role)
		require.NoError(t, genErr, "Pre-condition: Failed to generate token for validation test")

		// Validate generated token
		claims, err := jwtService.ValidateToken(validTokenString)

		require.NoError(t, err, "ValidateToken should not return an error for a valid token")
		require.NotNil(t, claims, "Claims should not be nil for a valid token")

		assert.Equal(t, username, claims.Username)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, username, claims.Subject)

	})

	t.Run("ValidateToken_Failure_MalformedToken", func(t *testing.T) {
		malformedTokenStrings := []string{
			"this.is.not.a.valid.jwt",                // Incorrect structure
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..", // Missing signature part
			"header.payload.",                        // Missing signature part
			"",                                       // Empty string
		}

		for _, tokenStr := range malformedTokenStrings {
			claims, err := jwtService.ValidateToken(tokenStr)
			require.Error(t, err, "ValidateToken should return an error for malformed token: %s", tokenStr)
			assert.Nil(t, claims, "Claims should be nil for malformed token")

			// The error from jwtService.ValidateToken is "invalid token: <original jwt error>"
			assert.Contains(t, err.Error(), "invalid token:", "Error meddsge should indicate an invalid token")

		}
	})

	t.Run("ValidateToken_Failure_InvalidSignature", func(t *testing.T) {
		// Using a signed key with a different secret
		differentSecret := []byte("a_completely_different_secret_key_!@#")

		standardClaims := domain.CustomClaims{
			Username: username,
			Role:     role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Subject:   username,
			},
		}
		tokenSignedWithWrongSecret := jwt.NewWithClaims(jwt.SigningMethodHS256, standardClaims)
		signedStringWithWrongSecret, signErr := tokenSignedWithWrongSecret.SignedString(differentSecret)
		require.NoError(t, signErr, "Fiaile dot sign token with a different secret for testing")

		// Validate the token
		claims, err := jwtService.ValidateToken(signedStringWithWrongSecret)
		require.Error(t, err, "ValidateToken shouild return an error for a token with an invalid signature")
		assert.Nil(t, claims, "Claims should be nil for invalid signature token")
		assert.Contains(t, err.Error(), jwt.ErrSignatureInvalid.Error(), "Error message should indicate invalid signature")

	})

	t.Run("ValidateToken_Failure_ExpiredToken", func(t *testing.T) {
		appSecret := []byte("ahnljdbjiohwebljnsknpihdbuo")
		expiredClaims := domain.CustomClaims{
			Username: username,
			Role:     role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Subject:   username,
			},
		}
		expiredTokenObject := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
		expiredTokenString, signErr := expiredTokenObject.SignedString(appSecret)
		require.NoError(t, signErr, "Failed to sign an expired token for testing")

		// Validate the expired token using the service
		claims, err := jwtService.ValidateToken(expiredTokenString)

		require.Error(t, err, "ValidateToken should return an error for an expired token")
		assert.Nil(t, claims, "Claims should be nil for an expired token")
		// The wrapped error should contain jwt.ErrTokenExpired
		assert.Contains(t, err.Error(), jwt.ErrTokenExpired.Error(), "Error message should indicate token is expired")

	})

	t.Run("ValidateToken_Failure_NotYetValidToken_NBF", func(t *testing.T) {
		// Generate a token that is not yet valid (due to NotBefore claim).
		appSecret := []byte("ahnljdbjiohwebljnsknpihdbuo")
		nbfClaims := domain.CustomClaims{
			Username: username,
			Role:     role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // Not valid for 1 hour
				Subject:   username,
			},
		}
		nbfTokenObject := jwt.NewWithClaims(jwt.SigningMethodHS256, nbfClaims)
		nbfTokenString, signErr := nbfTokenObject.SignedString(appSecret)
		require.NoError(t, signErr, "Failed to sign an NBF token for testing")

		// Validate the token
		claims, err := jwtService.ValidateToken(nbfTokenString)

		require.Error(t, err, "ValidateToken should return an error for a token that is not yet valid (NBF)")
		assert.Nil(t, claims, "Claims should be nil for an NBF token")
		// The wrapped error should contain jwt.ErrTokenNotValidYet
		assert.Contains(t, err.Error(), jwt.ErrTokenNotValidYet.Error(), "Error message should indicate token is not yet valid")
	})

	t.Run("ValidateToken_Failure_IncorrectSigningMethod", func(t *testing.T) {
		tokenWithAlgNone := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0ZXN0dXNlciJ9."

		valClaims, err := jwtService.ValidateToken(tokenWithAlgNone)
		require.Error(t, err, "ValidateToken should reject token with 'none' algorithm")
		assert.Nil(t, valClaims)
		assert.Contains(t, err.Error(), "unexpected signing method", "Error should be due to signing method check")
	})
}
