package utils_test

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomInt(t *testing.T) {
	Email := &utils.Utils{}
	result, err := Email.CreateVerificationString()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestSendVerificationEmail_InvalidCredentials(t *testing.T) {
	Email := &utils.Utils{}

	mockConfig := initializer.Config{
		ClientOrigin: "invalid-url",
		EmailAddress: "test@example.com",
		SMTPHost:     "sandbox.smtp.mailtrap.io",
		SMTPPort:     587,
		SMTPUser:     "test",
		SMTPPass:     "password",
	}
	utils.Config = mockConfig

	account := &models.Account{
		Email:                  "user@example.com",
		FirstName:              "John",
		EmailVerificationToken: "token",
	}

	err := Email.SendVerificationEmail(account)

	assert.Error(t, err)
}

func TestSendVerificationEmail_WrongUserEmailAddress(t *testing.T) {
	Email := &utils.Utils{}

	utils.Config = initializer.LoadProjConfig("../")

	account := &models.Account{
		FirstName:              "John",
		EmailVerificationToken: "token",
	}

	err := Email.SendVerificationEmail(account)

	assert.Error(t, err)
}

func TestSendVerificationEmail_WrongClientOrigin(t *testing.T) {
	Email := &utils.Utils{}

	mockConfig := initializer.Config{
		ClientOrigin: "asda   .com  ",
	}

	utils.Config = mockConfig

	account := &models.Account{
		Email:                  "user@example.com",
		FirstName:              "John",
		EmailVerificationToken: "token",
	}

	err := Email.SendVerificationEmail(account)

	assert.Error(t, err)
}
func TestSendVerificationEmail_Success(t *testing.T) {
	Email := &utils.Utils{}

	utils.Config = initializer.LoadProjConfig("../")

	account := &models.Account{
		Email:                  "user@example.com",
		FirstName:              "John",
		EmailVerificationToken: "token",
	}

	err := Email.SendVerificationEmail(account)

	assert.NoError(t, err)
}

func TestSendResetPassEmail_InvalidCredentials(t *testing.T) {
	Email := &utils.Utils{}
	mockConfig := initializer.Config{
		ClientOrigin: "invalid-url",
		EmailAddress: "test@example.com",
		SMTPHost:     "sandbox.smtp.mailtrap.io",
		SMTPPort:     587,
		SMTPUser:     "test",
		SMTPPass:     "password",
	}
	utils.Config = mockConfig

	account := &models.Account{
		Email:                  "user@example.com",
		FirstName:              "John",
		EmailVerificationToken: "token",
	}

	err := Email.SendVerificationEmail(account)

	assert.Error(t, err)
}

func TestSendResetPassEmail_success(t *testing.T) {
	Email := &utils.Utils{}
	utils.Config = initializer.LoadProjConfig("../")

	mockAccount := &models.Account{
		Email:                 "user@example.com",
		AccountPassResetToken: "token",
		FirstName:             "John",
	}

	err := Email.SendResetPassEmail(mockAccount)

	assert.NoError(t, err)
}
