package authcontrollers

import (
	"fmt"
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

		fmt.Println(p.Email)
		jwtToken, s_err := authservices.RequestNewAccount(p.Email)
		if s_err != nil {
			w_err := c.WriteJSON(map[string]any{"code": 400, "error": s_err})
			if w_err != nil {
				log.Println(w_err)
			}
			break
		}

		w_err := c.WriteJSON(map[string]any{"code": 200, "signup_session_jwt": jwtToken})
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
