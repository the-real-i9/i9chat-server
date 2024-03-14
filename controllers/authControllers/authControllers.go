package authcontrollers

import (
	"fmt"
	"i9chat/middlewares"
	"log"
	"services/authservices"
	"utils/mytypes"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var RequestNewAccount = websocket.New(func(c *websocket.Conn) {
	for {
		var body struct {
			Email string `json:"email"`
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		var w_err error

		jwtToken, app_err := authservices.RequestNewAccount(body.Email)

		if app_err != nil {
			w_err = c.WriteJSON(map[string]any{"code": fiber.StatusUnprocessableEntity, "error": app_err.Error()})
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
		if w_err := c.WriteJSON(map[string]any{"code": fiber.StatusUnprocessableEntity, "error": mid_err.Error()}); w_err != nil {
			log.Println(w_err)
			return
		}
		return
	}

	for {
		var body struct {
			Code int `json:"code"`
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		sessionData := sessData.(mytypes.SignupSessionData)

		app_err := authservices.VerifyEmail(sessionData.SessionId, body.Code, sessionData.Email)

		var w_err error

		if app_err != nil {
			w_err = c.WriteJSON(map[string]any{"code": fiber.StatusUnprocessableEntity, "error": app_err.Error()})
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
		if w_err := c.WriteJSON(map[string]any{"code": fiber.StatusUnprocessableEntity, "error": mid_err.Error()}); w_err != nil {
			log.Println(w_err)
			return
		}
		return
	}

	for {
		var body struct {
			Username    string `json:"username"`
			Password    string `json:"password"`
			Geolocation string `json:"geolocation"`
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		sessionData := sessData.(mytypes.SignupSessionData)

		userData, jwtToken, app_err := authservices.RegisterUser(sessionData.SessionId, sessionData.Email, body.Username, body.Password, body.Geolocation)

		var w_err error

		if app_err != nil {
			w_err = c.WriteJSON(map[string]any{"code": fiber.StatusUnprocessableEntity, "error": app_err.Error()})
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
	for {
		var body struct {
			EmailOrUsername string `json:"emailOrUsername"`
			Password        string `json:"password"`
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		userData, jwtToken, app_err := authservices.Signin(body.EmailOrUsername, body.Password)

		var w_err error

		if app_err != nil {
			w_err = c.WriteJSON(map[string]any{"code": fiber.StatusUnprocessableEntity, "error": app_err.Error()})
		} else {
			w_err = c.WriteJSON(map[string]any{"code": fiber.StatusOK, "msg": "Signin success!", "user": userData, "jwtToken": jwtToken})
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}
	}
})
