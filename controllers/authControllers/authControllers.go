package authcontrollers

import (
	"fmt"
	"i9chat/middlewares"
	"log"
	"services/authservices"
	"utils/apptypes"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var RequestNewAccount = websocket.New(func(c *websocket.Conn) {
	var body struct {
		Email string `json:"email"`
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		var w_err error

		jwtToken, app_err := authservices.RequestNewAccount(body.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{"code": fiber.StatusOK, "signup_session_jwt": jwtToken})
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}
	}
})

var VerifyEmail = websocket.New(func(c *websocket.Conn) {
	sessData, mid_err := middlewares.CheckAccountRequested(c)

	if mid_err != nil {
		if w_err := c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, mid_err)); w_err != nil {
			log.Println(w_err)
			return
		}
		return
	}

	var body struct {
		Code int `json:"code"`
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		sessionData := sessData.(apptypes.SignupSessionData)

		app_err := authservices.VerifyEmail(sessionData.SessionId, body.Code, sessionData.Email)

		var w_err error

		if app_err != nil {
			w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{"code": fiber.StatusOK, "msg": fmt.Sprintf("Your email '%s' has been verified!", sessionData.Email)})
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}
	}
})

var RegisterUser = websocket.New(func(c *websocket.Conn) {
	sessData, mid_err := middlewares.CheckEmailVerified(c)

	if mid_err != nil {
		if w_err := c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, mid_err)); w_err != nil {
			log.Println(w_err)
			return
		}
		return
	}

	var body struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		Geolocation string `json:"geolocation"`
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		sessionData := sessData.(apptypes.SignupSessionData)

		userData, jwtToken, app_err := authservices.RegisterUser(sessionData.SessionId, sessionData.Email, body.Username, body.Password, body.Geolocation)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{"code": fiber.StatusOK, "msg": "Signup success!", "user": userData, "jwtToken": jwtToken})
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}
	}
})

var Signin = websocket.New(func(c *websocket.Conn) {
	var body struct {
		EmailOrUsername string `json:"emailOrUsername"`
		Password        string `json:"password"`
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		userData, jwtToken, app_err := authservices.Signin(body.EmailOrUsername, body.Password)

		var w_err error

		if app_err != nil {
			w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{"code": fiber.StatusOK, "msg": "Signin success!", "user": userData, "jwtToken": jwtToken})
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}
	}
})
