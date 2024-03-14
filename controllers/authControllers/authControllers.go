package authcontrollers

import (
	"fmt"
	"i9chat/middlewares"
	"log"
	"services/authservices"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
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
			w_err = c.WriteJSON(map[string]any{"code": 400, "error": app_err.Error()})
		} else {
			w_err = c.WriteJSON(map[string]any{"code": 200, "signup_session_jwt": jwtToken})
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}
	}
})

var VerifyEmail = websocket.New(func(c *websocket.Conn) {
	for {
		var body struct {
			Code int `json:"code"`
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		sessData, mid_err := middlewares.CheckAccountRequested(c)

		if mid_err != nil {
			if w_err := c.WriteJSON(map[string]any{"code": 400, "error": mid_err.Error()}); w_err != nil {
				log.Println(w_err)
				break
			}
			break
		}

		var sessionData struct {
			SessionId string
			Email     string
		}

		helpers.MapToStruct(sessData, &sessionData)

		app_err := authservices.VerifyEmail(sessionData.SessionId, body.Code, sessionData.Email)

		var w_err error

		if app_err != nil {
			w_err = c.WriteJSON(map[string]any{"code": 400, "error": app_err.Error()})
		} else {
			w_err = c.WriteJSON(map[string]any{"code": 200, "msg": fmt.Sprintf("Your email '%s' has been verified!", sessionData.Email)})
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}
	}
})

var RegisterUser = websocket.New(func(c *websocket.Conn) {

})
