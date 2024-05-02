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

	err := Email.SendResetPassEmail(account)

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

func TestGenerateVerificationURL_Success(t *testing.T) {
	utils.Config.ClientOrigin = "www.test_domain.com"
	urlDetails := utils.URLConfig{
		ParamName: "test_param",
		Token:     "THI$_I$_TE$T_T@KEN",
		Path:      "/email_test_path",
	}

	link, err := utils.GenerateVerificationURL(urlDetails)

	assert.Equal(t, link, "www.test_domain.com/email_test_path?test_param=THI%24_I%24_TE%24T_T%40KEN")
	assert.NoError(t, err)

}
func TestGenerateVerificationURL_Fail(t *testing.T) {
	utils.Config.ClientOrigin = "www.test_domain.com"

	urlDetails := utils.URLConfig{
		ParamName: "",
		Token:     "THI$_I$_TE$T_T@KEN",
		Path:      "",
	}

	_, err := utils.GenerateVerificationURL(urlDetails)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid URL details provided")

}

func TestComposeEmail_IncompleteDetails(t *testing.T) {
	details := utils.EmailDetails{
		To:         "",
		Subject:    "Test Email",
		HTMLbody:   "",
		ButtonName: "",
	}

	err := utils.ComposeEmail(details)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "details are missing")
}
func TestComposeEmail_SMTPHostFail(t *testing.T) {

	mockConfig := initializer.Config{
		ClientOrigin: "exampleDomain.com",
		EmailAddress: "test@example.com",
		SMTPUser:     "test",
		SMTPPass:     "password",
	}

	utils.Config = mockConfig

	details := utils.EmailDetails{
		To:         "SomeEmail@exampleTest.com",
		UserName:   "ExampleName",
		Subject:    "Test Email",
		HTMLbody:   "<div>Just another element</div>",
		ButtonName: "Click Me",
	}

	utils.SMTPDetails.SMTPAuth = nil
	utils.SMTPDetails.SMTPHost = ""

	err := utils.ComposeEmail(details)

	assert.Error(t, err, "email was not sent successfully")
	assert.Contains(t, err.Error(), "missing address")
}

func TestComposeEmail_Success(t *testing.T) {
	FakeHostServer, port, serverStop := setupMockServer.MockSMTPServer()
	defer serverStop()

	mockConfig := initializer.Config{
		ClientOrigin: "exampleDomain.com",
		EmailAddress: "test@example.com",
		SMTPUser:     "test",
		SMTPPass:     "password",
	}

	utils.Config = mockConfig

	details := utils.EmailDetails{
		To:         "SomeEmail@exampleTest.com",
		UserName:   "ExampleName",
		Subject:    "Test Email",
		HTMLbody:   "<div>Just another element</div>",
		ButtonName: "Click Me",
	}

	utils.SMTPDetails.SMTPAuth = nil
	utils.SMTPDetails.SMTPHost = fmt.Sprintf("%s:%v", FakeHostServer, port)

	err := utils.ComposeEmail(details)

	assert.NoError(t, err, "email was sent successfully")

}
