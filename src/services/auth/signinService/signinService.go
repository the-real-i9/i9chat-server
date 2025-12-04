package signinService

import (
	"context"
	"i9chat/src/appErrors/userErrors"
	"i9chat/src/appTypes"
	user "i9chat/src/models/userModel"
	"i9chat/src/services/securityServices"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type signinRespT struct {
	Msg  string           `json:"msg"`
	User user.ToAuthUserT `json:"user"`
}

func Signin(ctx context.Context, emailOrUsername, password string) (signinRespT, string, error) {
	var resp signinRespT

	theUser, err := user.AuthFind(ctx, emailOrUsername)
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
		Username:      theUser.Username,
		ProfilePicUrl: theUser.ProfilePicUrl,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour))

	if err != nil {
		return resp, "", err
	}

	resp.Msg = "Signin success!"
	resp.User = theUser

	return resp, authJwt, nil
}
