package signinService

import (
	"context"
	"i9chat/src/appErrors/userErrors"
	"i9chat/src/appTypes"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/services/cloudStorageService"
	"i9chat/src/services/securityServices"
	"i9chat/src/services/userService"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

type signinRespT struct {
	Msg  string             `msgpack:"msg"`
	User UITypes.ClientUser `msgpack:"user"`
}

func Signin(ctx context.Context, emailOrUsername, password string) (signinRespT, string, error) {
	var resp signinRespT

	theUser, err := userService.SigninUserFind(ctx, emailOrUsername)
	if err != nil {
		return resp, "", err
	}

	if theUser.Username == "" {
		return resp, "", fiber.NewError(fiber.StatusNotFound, userErrors.IncorrectCredentials)
	}

	hashedPassword := theUser.Password

	yes, err := securityServices.PasswordMatchesHash(hashedPassword, password)
	if err != nil {
		return resp, "", err
	}

	if !yes {
		return resp, "", fiber.NewError(fiber.StatusNotFound, userErrors.IncorrectCredentials)
	}

	authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
		Username: theUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour))

	if err != nil {
		return resp, "", err
	}

	theUser.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(theUser.ProfilePicUrl)

	resp.Msg = "Signin success!"
	resp.User = UITypes.ClientUser{Username: theUser.Username, ProfilePicUrl: theUser.ProfilePicUrl, Presence: theUser.Presence}

	return resp, authJwt, nil
}
