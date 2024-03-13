package authcontrollers

import (
	"log"
	"services/authservices"

	"github.com/gofiber/contrib/websocket"
)

var RequestNewAccount = websocket.New(func(c *websocket.Conn) {
	for {
		var p struct {
			Email string `json:"email"`
		}

		r_err := c.ReadJSON(&p)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		jwtToken, app_err := authservices.RequestNewAccount(p.Email)

		var w_err error

		if app_err != nil {
			log.Println(app_err)
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

})

var RegisterUser = websocket.New(func(c *websocket.Conn) {

})
