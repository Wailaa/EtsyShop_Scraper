package utils

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"net/textproto"
	"net/url"

	"github.com/jordan-wright/email"
)

func CreateVerificationString() (string, error) {
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

func SendVerificationEmail(account *models.Account) error {

	Config := initializer.LoadProjConfig("/")

	verificationLink, err := url.Parse(Config.ClientOrigin)
	if err != nil {
		log.Fatal(err)
	}
	verificationLink.Path += "/verifyaccount"
	param := url.Values{}
	param.Add("TranID", account.EmailVerificationToken)
	verificationLink.RawQuery = param.Encode()

	e := &email.Email{
		To:      []string{account.Email},
		From:    Config.EmailAddress,
		Subject: "Confirm registration",
		Text:    []byte("Text Body is, of course, supported!"),
		HTML: []byte(fmt.Sprintf(`<html>
		<head>
		<div>
		<h1>Hello %s,</h1>
		<p>We’re happy you signed up for EtsyScraper. To start reading shops data , please confirm your email address.</p>
		<button>
		<a href="%s">Verify Now</a>
		</button>
		<p>Welcome to EtsyScraper!</p>
		<p>The EtsyScraper Team</p>
	  	</div>
	 </body>
		</html>`, account.FirstName, verificationLink.String())),
		Headers: textproto.MIMEHeader{},
	}

	err = e.Send(fmt.Sprintf("%s:%v", Config.SMTPHost, Config.SMTPPort), smtp.PlainAuth(Config.EmailAddress, Config.SMTPUser, Config.SMTPPass, Config.SMTPHost))
	if err != nil {
		fmt.Println("There was an error sending the mail", err)
	}
	return nil
}

func SendResetPassEmail(account *models.Account) error {

	Config := initializer.LoadProjConfig("/")

	verificationLink, err := url.Parse(Config.ClientOrigin)
	if err != nil {
		log.Fatal(err)
	}
	verificationLink.Path += "/reset_password"
	param := url.Values{}
	param.Add("rcp", account.AccountPassResetToken)
	verificationLink.RawQuery = param.Encode()

	e := &email.Email{
		To:      []string{account.Email},
		From:    Config.EmailAddress,
		Subject: "Reset Password",
		Text:    []byte("This is an email sent upon your request to reset your account password at EtsyScraper. if you did not request this email then please igonre the message, otherwise please press on the button to reset your password"),
		HTML: []byte(fmt.Sprintf(`<html>
		<head>
		<div>
		<h1>Hello %s,</h1>
		<p>This is an email sent upon your request to reset your account password at EtsyScraper </p>
		<p>if you did not request this email then please igonre the message, otherwise please press on the button to reset your password.  </p>
		<button>
		<a href="%s">Reset Password</a>
		</button>
		<p>The EtsyScraper Team</p>
	  	</div>
	 </body>
		</html>`, account.FirstName, verificationLink.String())),
		Headers: textproto.MIMEHeader{},
	}

	err = e.Send(fmt.Sprintf("%s:%v", Config.SMTPHost, Config.SMTPPort), smtp.PlainAuth(Config.EmailAddress, Config.SMTPUser, Config.SMTPPass, Config.SMTPHost))
	if err != nil {
		log.Println("There was an error sending the mail", err)
		return err
	}
	return nil
}
