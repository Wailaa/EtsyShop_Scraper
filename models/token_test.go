package models_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"EtsyScraper/models"
)

func TestNewTokenWithValidToken(t *testing.T) {
	token := "valid_token"
	newToken := models.NewToken(token)

	if newToken == nil {
		t.Errorf("Expected newToken to not be nil")
	}

	if newToken != nil {

		if models.Token(token) != *newToken {
			t.Errorf("Expected newToken to be %s, but got %s", token, *newToken)
		}
	}
}

func TestCreateClaimsValidIClaims(t *testing.T) {

	claims := jwt.MapClaims{
		"iat":      float64(1234567890),
		"exp":      float64(1234567890),
		"userUUID": "123e4567-e89b-12d3-a456-426614174000",
	}

	customClaims := models.CreateClaims(claims)

	expectedCreatedAt := int64(1234567890)
	expectedExpiresAt := int64(1234567890)
	expectedUserUUID, _ := uuid.Parse("123e4567-e89b-12d3-a456-426614174000")

	assert.Equal(t, expectedCreatedAt, customClaims.CreatedAt)
	assert.Equal(t, expectedExpiresAt, customClaims.ExpiresAt)
	assert.Equal(t, expectedUserUUID, customClaims.UserUUID)

}
