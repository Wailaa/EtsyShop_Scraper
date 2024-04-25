package utils_test

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
	"EtsyScraper/utils"
	"fmt"
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
		SMTPHost:     "sandbox.smtp.mailtrp.io",
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

	mockConfig := initializer.Config{
		ClientOrigin: "test.com",
	}

	utils.Config = mockConfig

	account := &models.Account{
		FirstName:              "John",
		EmailVerificationToken: "token",
	}

	err := Email.SendVerificationEmail(account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mail: no address")
}

func TestSendVerificationEmail_WrongClientOrigin(t *testing.T) {
	Email := &utils.Utils{}

	mockConfig := initializer.Config{
		ClientOrigin: "asda .com  ",
		EmailAddress: "Test@Test.com",
	}

	utils.Config = mockConfig

	account := &models.Account{
		Email:                  "user@example.com",
		FirstName:              "John",
		EmailVerificationToken: "token",
	}

	err := Email.SendVerificationEmail(account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can't assign requested address")

}
func TestSendVerificationEmail_Success(t *testing.T) {
	FakeHostServer, port, serverStop := setupMockServer.MockSMTPServer()
	defer serverStop()

	Email := &utils.Utils{}

	mockConfig := initializer.Config{
		ClientOrigin: "asda.com",
		EmailAddress: "Test@testing.com",
		SMTPUser:     "TestOnly",
		SMTPPass:     "NotYour1234",
		SMTPHost:     FakeHostServer,
		SMTPPort:     port,
	}

	utils.Config = mockConfig

	utils.SMTPDetails.SMTPAuth = nil
	utils.SMTPDetails.SMTPHost = fmt.Sprintf("%s:%v", FakeHostServer, port)

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
		SMTPHost:     "fakeHost",
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

	FakeHostServer, port, serverStop := setupMockServer.MockSMTPServer()
	defer serverStop()

	Email := &utils.Utils{}
	mockConfig := initializer.Config{
		ClientOrigin: "exampleDomain.com",
		EmailAddress: "test@example.com",
		SMTPUser:     "test",
		SMTPPass:     "password",
	}

	utils.Config = mockConfig
	utils.SMTPDetails.SMTPAuth = nil
	utils.SMTPDetails.SMTPHost = fmt.Sprintf("%s:%v", FakeHostServer, port)

	mockAccount := &models.Account{
		Email:                 "user@example.com",
		AccountPassResetToken: "token",
		FirstName:             "John",
	}

	err := Email.SendResetPassEmail(mockAccount)

	assert.NoError(t, err)
}
