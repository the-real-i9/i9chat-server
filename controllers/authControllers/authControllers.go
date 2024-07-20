package authControllers

import (
	"fmt"
	"i9chat/middlewares"
	"i9chat/services/authServices"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var RequestNewAccount = websocket.New(func(c *websocket.Conn) {
	var body struct {
		Email string `validate:"required,email"`
	}

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := helpers.Validate.Struct(body); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		signupSessionJwt, app_err := authServices.RequestNewAccount(body.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":              "A 6-digit verification code has been sent to " + body.Email,
				"signupSessionJwt": signupSessionJwt,
			},
		})
	}
})

var VerifyEmail = websocket.New(func(c *websocket.Conn) {
	sessionData, mid_err := middlewares.CheckAccountRequested(c)

	if mid_err != nil {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, mid_err)); w_err != nil {
			log.Println(w_err)
			return
		}
		return
	}

	var body struct {
		Code int `validate:"required,min=6"`
	}

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := helpers.Validate.Struct(body); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		app_err := authServices.VerifyEmail(sessionData.SessionId, body.Code, sessionData.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg": fmt.Sprintf("Your email '%s' has been verified!", sessionData.Email),
			},
		})
	}
})

var RegisterUser = websocket.New(func(c *websocket.Conn) {
	sessionData, mid_err := middlewares.CheckEmailVerified(c)

	if mid_err != nil {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, mid_err)); w_err != nil {
			log.Println(w_err)
			return
		}
		return
	}

	var body struct {
		Username    string `validate:"required,min=3,alphanumunicode"`
		Password    string `validate:"required,min=8"`
		Geolocation string `validate:"required"`
	}

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := helpers.Validate.Struct(body); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		userData, authJwt, app_err := authServices.RegisterUser(sessionData.SessionId, sessionData.Email, body.Username, body.Password, body.Geolocation)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":     "Signup success!",
				"user":    userData,
				"authJwt": authJwt,
			},
		})
	}
})

var Signin = websocket.New(func(c *websocket.Conn) {
	var body struct {
		EmailOrUsername string `validate:"required,email|alphanumunicode,min=6"`
		Password        string `validate:"required"`
	}

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := helpers.Validate.Struct(body); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		userData, authJwt, app_err := authServices.Signin(body.EmailOrUsername, body.Password)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":     "Signin success!",
				"user":    userData,
				"authJwt": authJwt,
			},
		})
	}
})
