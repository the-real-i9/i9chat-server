package authControllers

import (
	"fmt"
	"i9chat/appTypes"
	user "i9chat/models/userModel"
	"i9chat/services/authServices"
	"os"
	"time"
)

func signin(emailOrUsername string, password string) (*user.User, string, error) {
	theUser, err := user.FindOne(emailOrUsername)
	if err != nil {
		return nil, "", err
	}

	if theUser == nil {
		return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
	}

	hashedPassword, err := user.GetPassword(emailOrUsername)
	if err != nil {
		return nil, "", err
	}

	if !authServices.PasswordMatchesHash(hashedPassword, password) {
		return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
	}

	authJwt := authServices.JwtSign(appTypes.ClientUser{
		Id:       theUser.Id,
		Username: theUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour))

	return theUser, authJwt, nil
}
