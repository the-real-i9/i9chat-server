package middlewares

import (
	"fmt"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
)

var nilSsd = appTypes.SignupSessionData{}

func CheckAccountRequested(c *websocket.Conn) (appTypes.SignupSessionData, error) {
	token := c.Headers("Authorization")

	if token == "" {
		return nilSsd, fmt.Errorf("signup error: no ongoing signup session. you must first submit your email and attach the autorization token sent")
	}

	sessData, err := helpers.JwtVerify(token, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
	if err != nil {
		if err.Error() == "authorization error: invalid jwt" {
			return nilSsd, fmt.Errorf("signup error: invalid signup session token")
		}
		if err.Error() == "authorization error: jwt expired" {
			return nilSsd, fmt.Errorf("signup error: signup session expired")
		}
	}

	var sessionData appTypes.SignupSessionData

	helpers.ParseToStruct(sessData, &sessionData)

	return sessionData, nil
}

func CheckEmailVerified(c *websocket.Conn) (appTypes.SignupSessionData, error) {
	token := c.Headers("Authorization")

	if token == "" {
		return nilSsd, fmt.Errorf("signup error: no ongoing signup session. you must first submit your email and attach the autorization token sent")
	}

	sessData, err := helpers.JwtVerify(token, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
	if err != nil {
		if err.Error() == "authorization error: invalid jwt" {
			return nilSsd, fmt.Errorf("signup error: invalid signup session token")
		}
		if err.Error() == "authorization error: jwt expired" {
			return nilSsd, fmt.Errorf("signup error: signup session expired")
		}
	}

	var sessionData appTypes.SignupSessionData

	helpers.ParseToStruct(sessData, &sessionData)

	isVerified, err := helpers.QueryRowField[bool]("SELECT is_verified FROM signup_session_email_verified($1)", sessionData.SessionId)
	if err != nil {
		log.Println(fmt.Errorf("middlewares: CheckEmailVerified: isVerified: db error: %s", err))
		return nilSsd, helpers.ErrInternalServerError
	}

	if !*isVerified {
		return nilSsd, fmt.Errorf("signup error: your email '%s' has not been verified", sessionData.Email)
	}

	return sessionData, nil
}
