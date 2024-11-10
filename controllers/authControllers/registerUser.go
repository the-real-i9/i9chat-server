package authControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/models/appModel"
	user "i9chat/models/userModel"
	"i9chat/services/authServices"
	"os"
	"time"
)

func registerUser(sessionId, email, username, password, geolocation string) (*user.User, string, error) {
	accExists, err := appModel.AccountExists(username)
	if err != nil {
		return nil, "", err
	}

	if accExists {
		return nil, "", fmt.Errorf("username error: username '%s' is unavailable", username)
	}

	hashedPassword := authServices.HashPassword(password)

	newUser, err := user.New(email, username, hashedPassword, geolocation)
	if err != nil {
		return nil, "", err
	}

	authJwt := authServices.JwtSign(appTypes.ClientUser{
		Id:       newUser.Id,
		Username: newUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour)) // 1 year

	go appModel.EndSignupSession(sessionId)

	return newUser, authJwt, nil
}
