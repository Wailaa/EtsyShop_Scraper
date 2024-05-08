package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"net/textproto"
	"net/url"

	"github.com/jordan-wright/email"

	initializer "EtsyScraper/init"
	"EtsyScraper/models"
)

var Config = initializer.LoadProjConfig(".")

type EmailConfig struct {
	SMTPHost string
	SMTPAuth smtp.Auth
}

type URLConfig struct {
	ParamName string
	Token     string
	Path      string
}

type EmailDetails struct {
	To               string
	UserName         string
	Subject          string
	HTMLbody         string
	ButtonName       string
	Plaintext        string
	VerificationLink string
}

var SMTPDetails = new(EmailConfig)

func init() {
	SMTPDetails.SMTPHost = fmt.Sprintf("%s:%v", Config.SMTPHost, Config.SMTPPort)
	SMTPDetails.SMTPAuth = smtp.PlainAuth(Config.EmailAddress, Config.SMTPUser, Config.SMTPPass, Config.SMTPHost)
}

func (em *Utils) CreateVerificationString() (string, error) {
	GenerateRandomInt, err := rand.Int(rand.Reader, big.NewInt(20))
	if err != nil {
		return "", err
	}

	byteLength := GenerateRandomInt.Int64() + 10
	CreateBytes := make([]byte, byteLength)

	if _, err := rand.Read(CreateBytes); err != nil {
		return "", err
	}

	EncodedString := base64.StdEncoding.EncodeToString(CreateBytes)
	return EncodedString, nil
}

func (em *Utils) SendVerificationEmail(account *models.Account) error {

	urlDetails := URLConfig{
		ParamName: "TranID",
		Token:     account.EmailVerificationToken,
		Path:      "/verify_account",
	}
	verificationLink, err := GenerateVerificationURL(urlDetails)
	if err != nil {
		return err
	}
	PlainText := "This is an email sent upon your registration at EtsyScraper."

	HTMLbody := `<p>We’re happy you signed up for EtsyScraper. To start reading shops data , please confirm your email address.</p>`

	details := EmailDetails{
		To:               account.Email,
		UserName:         account.FirstName,
		Subject:          "Confirm registration",
		Plaintext:        PlainText,
		HTMLbody:         HTMLbody,
		ButtonName:       "Verify Now",
		VerificationLink: verificationLink,
	}

	err = ComposeEmail(details)
	if err != nil {
		return fmt.Errorf("failed to send verification mail: no address or missig details error: %w", err)
	}
	return nil
}

func (em *Utils) SendResetPassEmail(account *models.Account) error {

	urlDetails := URLConfig{
		ParamName: "rcp",
		Token:     account.AccountPassResetToken,
		Path:      "/reset_password",
	}
	verificationLink, err := GenerateVerificationURL(urlDetails)
	if err != nil {
		return err
	}

	PlainText := "This is an email sent upon your request to reset your account password at EtsyScraper. if you did not request this email then please igonre the message, otherwise please press on the button to reset your password"

	HTMLbody := `<p>This is an email sent upon your request to reset your account password at EtsyScraper </p>
	<p>if you did not request this email then please igonre the message, otherwise please press on the button to reset your password.  </p>
	<button>`

	details := EmailDetails{
		To:               account.Email,
		UserName:         account.FirstName,
		Subject:          "Reset Password",
		Plaintext:        PlainText,
		HTMLbody:         HTMLbody,
		ButtonName:       "Reset Password",
		VerificationLink: verificationLink,
	}

	err = ComposeEmail(details)

	if err != nil {
		return fmt.Errorf("failed to send verification email,error: %w", err)
	}
	return nil
}

func GenerateVerificationURL(urlDetails URLConfig) (string, error) {

	if urlDetails.Path == "" || urlDetails.ParamName == "" || urlDetails.Token == "" {
		return "", errors.New("invalid URL details provided")
	}

	verificationLink, err := url.Parse(Config.ClientOrigin)
	if err != nil {
		return "", err
	}

	verificationLink.Path += urlDetails.Path
	param := url.Values{}
	param.Add(urlDetails.ParamName, urlDetails.Token)
	verificationLink.RawQuery = param.Encode()

	return verificationLink.String(), nil
}

func ComposeEmail(details EmailDetails) error {

	if details.To == "" || details.Subject == "" || details.HTMLbody == "" || details.ButtonName == "" || details.UserName == "" {
		return errors.New("could no compose email , details are missing")
	}

	e := &email.Email{
		To:      []string{details.To},
		From:    Config.EmailAddress,
		Subject: details.Subject,
		Text:    []byte(details.Plaintext),
		HTML: []byte(fmt.Sprintf(`<html>
		<head>
		<div>
		<h1>Hello %s,</h1>
		%s
		<button>
		<a href="%s">%s</a>
		</button>
		<p>The EtsyScraper Team</p>
		</div>
		</body>
		</html>`, details.UserName, details.HTMLbody, details.VerificationLink, details.ButtonName)),
		Headers: textproto.MIMEHeader{},
	}
	err := e.Send(SMTPDetails.SMTPHost, SMTPDetails.SMTPAuth)
	if err != nil {
		log.Println("error while composing email", err)
		return err
	}
	return nil
}
