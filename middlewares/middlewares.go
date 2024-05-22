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
	signupSessionJwt := c.Headers("Authorization")

	if signupSessionJwt == "" {
		return nilSsd, fmt.Errorf("authorization error: authorization token required")
	}

	sessData, err := helpers.JwtVerify(signupSessionJwt, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
	if err != nil {
		return nilSsd, err
	}

	var sessionData appTypes.SignupSessionData

	helpers.ParseToStruct(sessData, &sessionData)

	return sessionData, nil
}

func CheckEmailVerified(c *websocket.Conn) (appTypes.SignupSessionData, error) {
	signupSessionJwt := c.Headers("Authorization")

	if signupSessionJwt == "" {
		return nilSsd, fmt.Errorf("authorization error: authorization token required")
	}

	sessData, err := helpers.JwtVerify(signupSessionJwt, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
	if err != nil {
		return nilSsd, err
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
