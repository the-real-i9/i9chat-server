package middlewares

import (
	"fmt"
	"log"
	"os"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
)

func CheckAccountRequested(c *websocket.Conn) (map[string]any, error) {
	token := c.Headers("Authorization")

	if token == "" {
		return nil, fmt.Errorf("signup error: no ongoing signup session. you must first submit your email and attach the autorization token sent")
	}

	sessionData, err := helpers.JwtParse(token, os.Getenv("SIGNUP_JWT_SECRET"))
	if err != nil {
		if err.Error() == "authentication error: invalid jwt" {
			return nil, fmt.Errorf("signup error: invalid signup session token")
		}
		if err.Error() == "authentication error: jwt expired" {
			return nil, fmt.Errorf("signup error: signup session expired")
		}
	}

	return sessionData, nil
}

func CheckEmailVerified(c *websocket.Conn) error {
	token := c.Headers("Authorization")

	if token == "" {
		return fmt.Errorf("signup error: no ongoing signup session. you must first submit your email and attach the autorization token sent")
	}

	sessData, err := helpers.JwtParse(token, os.Getenv("SIGNUP_JWT_SECRET"))
	if err != nil {
		if err.Error() == "authentication error: invalid jwt" {
			return fmt.Errorf("signup error: invalid signup session token")
		}
		if err.Error() == "authentication error: jwt expired" {
			return fmt.Errorf("signup error: signup session expired")
		}
	}

	var sessionData struct {
		SessionId string
		Email     string
	}

	helpers.MapToStruct(sessData, &sessionData)

	verified, err := helpers.QueryRowField[bool]("SELECT verified FROM ongoing_signup WHERE session_id = $1", sessionData.SessionId)
	if err != nil {
		log.Println(err)
		return err
	}

	if !*verified {
		return fmt.Errorf("signup error: your email '%s' has not been verified", sessionData.Email)
	}

	return nil
}
