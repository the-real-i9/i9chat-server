package authControllers

import (
	"fmt"
	"i9chat/middlewares"
	"i9chat/services/authServices"
	"i9chat/utils/helpers"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var RequestNewAccount = websocket.New(func(c *websocket.Conn) {
	var body struct {
		Email string
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			// log.Println(r_err)
			break
		}

		signupSessionJwt, app_err := authServices.RequestNewAccount(body.Email)

		var w_err error

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": fiber.StatusOK,
				"body": map[string]any{
					"msg":              "A 6-digit verification code has been sent to " + body.Email,
					"signupSessionJwt": signupSessionJwt,
				},
			})
		}

		if w_err != nil {
			// log.Println(w_err)
			break
		}
	}
})

var VerifyEmail = websocket.New(func(c *websocket.Conn) {
	sessionData, mid_err := middlewares.CheckAccountRequested(c)

	if mid_err != nil {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, mid_err)); w_err != nil {
			// log.Println(w_err)
			return
		}
		return
	}

	var body struct {
		Code int
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			// log.Println(r_err)
			break
		}

		app_err := authServices.VerifyEmail(sessionData.SessionId, body.Code, sessionData.Email)

		var w_err error

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": fiber.StatusOK,
				"body": map[string]any{
					"msg": fmt.Sprintf("Your email '%s' has been verified!", sessionData.Email),
				},
			})
		}

		if w_err != nil {
			// log.Println(w_err)
			break
		}
	}
})

var RegisterUser = websocket.New(func(c *websocket.Conn) {
	sessionData, mid_err := middlewares.CheckEmailVerified(c)

	if mid_err != nil {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, mid_err)); w_err != nil {
			// log.Println(w_err)
			return
		}
		return
	}

	var body struct {
		Username    string
		Password    string
		Geolocation string
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			// log.Println(r_err)
			break
		}

		userData, authJwt, app_err := authServices.RegisterUser(sessionData.SessionId, sessionData.Email, body.Username, body.Password, body.Geolocation)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": fiber.StatusOK,
				"body": map[string]any{
					"msg":     "Signup success!",
					"user":    userData,
					"authJwt": authJwt,
				},
			})
		}

		if w_err != nil {
			// log.Println(w_err)
			break
		}
	}
})

var Signin = websocket.New(func(c *websocket.Conn) {
	var body struct {
		EmailOrUsername string
		Password        string
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			// log.Println(r_err)
			break
		}

		userData, authJwt, app_err := authServices.Signin(body.EmailOrUsername, body.Password)

		var w_err error

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": fiber.StatusOK,
				"body": map[string]any{
					"msg":     "Signin success!",
					"user":    userData,
					"authJwt": authJwt,
				},
			})
		}

		if w_err != nil {
			// log.Println(w_err)
			break
		}
	}
})
