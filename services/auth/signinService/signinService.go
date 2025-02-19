package signinService

import (
	"context"
	"i9chat/appTypes"
	user "i9chat/models/userModel"
	"i9chat/services/securityServices"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Signin(ctx context.Context, emailOrUsername, password string) (any, string, error) {
	theUser, err := user.SigninFind(ctx, emailOrUsername)
	if err != nil {
		return nil, "", err
	}

	if theUser == nil {
		return nil, "", fiber.NewError(fiber.StatusNotFound, "signin error: incorrect email/username or password")
	}

	hashedPassword := theUser["password"].(string)

	yes, err := securityServices.PasswordMatchesHash(hashedPassword, password)
	if err != nil {
		return nil, "", err
	}

	if !yes {
		return nil, "", fiber.NewError(fiber.StatusNotFound, "signin error: incorrect email/username or password")
	}

	authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
		Username: theUser["username"].(string),
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour))

	if err != nil {
		return nil, "", err
	}

	delete(theUser, "password")

	respData := map[string]any{
		"msg":  "Signin success!",
		"user": theUser,
	}

	return respData, authJwt, nil
}
