package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPass(pass string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashed)

}
