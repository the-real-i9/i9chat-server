package signupControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/services/auth/signupService"
	"i9chat/services/utils/authUtilServices"
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

		respData, app_err := signupService.RequestNewAccount(body.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       respData,
		})
	}
})

var VerifyEmail = websocket.New(func(c *websocket.Conn) {
	sessionToken := c.Headers("Authorization")

	sessionData, err := authUtilServices.JwtVerify[appTypes.SignupSessionData](sessionToken, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
	if err != nil {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err)); w_err != nil {
			log.Println(w_err)
		}
		return
	}

	if sessionData.Step != "verify email" {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, fmt.Errorf("invalid session token on endpoint"))); w_err != nil {
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

		respData, app_err := signupService.VerifyEmail(sessionData.SessionId, body.Code, sessionData.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       respData,
		})
	}
})

var RegisterUser = websocket.New(func(c *websocket.Conn) {
	sessionToken := c.Headers("Authorization")

	sessionData, err := authUtilServices.JwtVerify[appTypes.SignupSessionData](sessionToken, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
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

		respData, app_err := signupService.RegisterUser(sessionData.SessionId, sessionData.Email, body.Username, body.Password, body.Geolocation)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       respData,
		})
	}
})
