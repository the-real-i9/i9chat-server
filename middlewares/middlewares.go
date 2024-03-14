package middlewares

import (
	"fmt"
	"log"
	"os"
	"utils/helpers"
	"utils/mytypes"

	"github.com/gofiber/contrib/websocket"
)

func CheckAccountRequested(c *websocket.Conn) (any, error) {
	token := c.Headers("Authorization")

	if token == "" {
		return nil, fmt.Errorf("signup error: no ongoing signup session. you must first submit your email and attach the autorization token sent")
	}

	sessData, err := helpers.JwtVerify(token, os.Getenv("SIGNUP_JWT_SECRET"))
	if err != nil {
		if err.Error() == "authentication error: invalid jwt" {
			return nil, fmt.Errorf("signup error: invalid signup session token")
		}
		if err.Error() == "authentication error: jwt expired" {
			return nil, fmt.Errorf("signup error: signup session expired")
		}
	}

	var sessionData mytypes.SignupSessionData

	helpers.MapToStruct(sessData, &sessionData)

	return sessionData, nil
}

func CheckEmailVerified(c *websocket.Conn) (any, error) {
	token := c.Headers("Authorization")

	if token == "" {
		return nil, fmt.Errorf("signup error: no ongoing signup session. you must first submit your email and attach the autorization token sent")
	}

	sessData, err := helpers.JwtVerify(token, os.Getenv("SIGNUP_JWT_SECRET"))
	if err != nil {
		if err.Error() == "authorization error: invalid jwt" {
			return nil, fmt.Errorf("signup error: invalid signup session token")
		}
		if err.Error() == "authorization error: jwt expired" {
			return nil, fmt.Errorf("signup error: signup session expired")
		}
	}

	var sessionData mytypes.SignupSessionData

	helpers.MapToStruct(sessData, &sessionData)

	isVerified, err := helpers.QueryRowField[bool]("SELECT is_verified FROM signup_session_email_verified($1)", sessionData.SessionId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !*isVerified {
		return nil, fmt.Errorf("signup error: your email '%s' has not been verified", sessionData.Email)
	}

	return sessionData, nil
}
