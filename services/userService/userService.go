package userservice

import (
	"model/usermodel"
	"os"
	"time"
	"utils/helpers"
)

func NewUser(email string, username string, password string, geolocation string) (map[string]any, string, error) {
	var hashedPassword string

	// hash password here

	userData, err := usermodel.NewUser(email, username, hashedPassword, geolocation)
	if err != nil {
		return nil, "", err
	}

	var user struct {
		Id       int
		Username string
	}

	helpers.MapToStruct(userData, &user)

	// sign a jwt token
	jwtToken := helpers.JwtSign(map[string]any{
		"userId":   user.Id,
		"username": user.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour)) // 1 year

	return userData, jwtToken, nil
}
