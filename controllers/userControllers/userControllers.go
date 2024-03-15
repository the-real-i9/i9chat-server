package usercontrollers

import (
	"utils/apptypes"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
)

var GetMyChats = helpers.WSHandlerProtected(func(c *websocket.Conn) {

	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

})
