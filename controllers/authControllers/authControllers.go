package authControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/services/authServices"
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var RequestNewAccount = websocket.New(func(c *websocket.Conn) {

	var w_err error

	for {
		var body requestNewAccountBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {

			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		signupSessionJwt, app_err := requestNewAccount(body.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":           "A 6-digit verification code has been sent to " + body.Email,
				"session_token": signupSessionJwt,
			},
		})
	}
})

var VerifyEmail = websocket.New(func(c *websocket.Conn) {
	sessionToken := c.Headers("Authorization")

	sessionData, err := authServices.JwtVerify[appTypes.SignupSessionData](sessionToken, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
	if err != nil {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err)); w_err != nil {
			log.Println(w_err)
		}
		return
	}

	if sessionData.Step != "verify email" {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, fmt.Errorf("expected state: verify email"))); w_err != nil {
			log.Println(w_err)
		}
		return
	}

	var w_err error

	for {
		var body verifyEmailBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		signupSessionJwt, app_err := verifyEmail(sessionData.SessionId, body.Code, sessionData.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":           fmt.Sprintf("Your email '%s' has been verified!", sessionData.Email),
				"session_token": signupSessionJwt,
			},
		})
	}
})

var RegisterUser = websocket.New(func(c *websocket.Conn) {
	sessionToken := c.Headers("Authorization")

	sessionData, err := authServices.JwtVerify[appTypes.SignupSessionData](sessionToken, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
	if err != nil {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err)); w_err != nil {
			log.Println(w_err)
		}
		return
	}

	if sessionData.Step != "register user" {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, fmt.Errorf("expected state: register user"))); w_err != nil {
			log.Println(w_err)
		}
		return
	}

	var w_err error

	for {
		var body registerUserBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		newUser, authJwt, app_err := registerUser(sessionData.SessionId, sessionData.Email, body.Username, body.Password, body.Geolocation)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":     "Signup success!",
				"user":    newUser,
				"authJwt": authJwt,
			},
		})
	}
})

var Signin = websocket.New(func(c *websocket.Conn) {

	var w_err error

	for {
		var body signInBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		theUser, authJwt, app_err := signin(body.EmailOrUsername, body.Password)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":     "Signin success!",
				"user":    theUser,
				"authJwt": authJwt,
			},
		})
	}
})
