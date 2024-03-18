package chatcontrollers

import (
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
)

var GetDMChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {

})
