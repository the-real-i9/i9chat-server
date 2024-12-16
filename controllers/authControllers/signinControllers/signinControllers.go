package signinControllers

import (
	"context"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/services/auth/signinService"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var Signin = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
			w_err = c.WriteJSON(helpers.ErrResp(val_err))
			continue
		}

		respData, app_err := signinService.Signin(ctx, body.EmailOrUsername, body.Password)

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
