package signinService

import (
	"fmt"
	"i9chat/appTypes"
	user "i9chat/models/userModel"
	"i9chat/services/securityServices"
	"os"
	"time"
)

func Signin(emailOrUsername string, password string) (any, error) {
	theUser, err := user.FindOne(emailOrUsername)
	if err != nil {
		return nil, err
	}

	if theUser == nil {
		return nil, fmt.Errorf("signin error: incorrect email/username or password")
	}

	hashedPassword, err := user.GetPassword(emailOrUsername)
	if err != nil {
		return nil, err
	}

	if !securityServices.PasswordMatchesHash(hashedPassword, password) {
		return nil, fmt.Errorf("signin error: incorrect email/username or password")
	}

	authJwt := securityServices.JwtSign(appTypes.ClientUser{
		Id:       theUser.Id,
		Username: theUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour))

	respData := map[string]any{
		"msg":     "Signin success!",
		"user":    theUser,
		"authJwt": authJwt,
	}

	return respData, nil
}
