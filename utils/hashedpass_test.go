package utils_test

import (
	"EtsyScraper/utils"
	"testing"
)

func TestValidPassword(t *testing.T) {
	password := "password123"
	hashedPass, err := utils.HashPass(password)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if hashedPass == "" {
		t.Errorf("Expected non-empty string, got empty string")
	}
}

func TestCorrectPassword(t *testing.T) {
	pass := "password123"
	hashedPass, _ := utils.HashPass(pass)

	result := utils.IsPassVerified(pass, hashedPass)

	if !result {
		t.Errorf("Expected true, but got false")
	}
}

func TestCorrectPassword_NoPassMatch(t *testing.T) {
	pass := "password123"
	wrongPass := "123password"
	hashedPass, _ := utils.HashPass(pass)

	result := utils.IsPassVerified(wrongPass, hashedPass)

	if result {
		t.Errorf("Expected false, but got true")
	}
}
