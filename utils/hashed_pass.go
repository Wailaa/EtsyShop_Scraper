package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func (pw *Utils) HashPass(pass string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", HandleError(err)
	}
	return string(hashed), nil
}

func (pw *Utils) IsPassVerified(pass string, hashedPass string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(pass)); err != nil {
		return false
	}
	return true
}
