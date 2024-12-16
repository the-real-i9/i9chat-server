package signupControllers

import (
	"context"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/services/auth/signupService"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var RequestNewAccount = websocket.New(func(c *websocket.Conn) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

			w_err = c.WriteJSON(helpers.ErrResp(val_err))
			continue
		}

		respData, app_err := signupService.RequestNewAccount(ctx, body.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       respData,
		})
	}
})

var VerifyEmail = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionData := c.Locals("session").(*appTypes.SignupSessionData)

	if sessionData.Step != "verify email" {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.NewError(fiber.StatusUnauthorized, "invalid session token on endpoint"))); w_err != nil {
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
			w_err = c.WriteJSON(helpers.ErrResp(val_err))
			continue
		}

		respData, app_err := signupService.VerifyEmail(ctx, sessionData.SessionId, body.Code, sessionData.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       respData,
		})
	}
})

var RegisterUser = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionData := c.Locals("session").(*appTypes.SignupSessionData)

	if sessionData.Step != "register user" {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.NewError(fiber.StatusUnauthorized, "invalid session token on endpoint"))); w_err != nil {
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
			w_err = c.WriteJSON(helpers.ErrResp(fiber.NewError(fiber.StatusBadRequest, "validation error:", val_err.Error())))
			continue
		}

		respData, app_err := signupService.RegisterUser(ctx, sessionData.SessionId, sessionData.Email, body.Username, body.Password, body.Geolocation)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       respData,
		})
	}
})
