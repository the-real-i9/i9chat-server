package chatcontrollers

import (
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
)

var ListenForNewMessage = helpers.WSHandlerProtected(func(c *websocket.Conn) {

})
