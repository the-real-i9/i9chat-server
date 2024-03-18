package chatcontrollers

import (
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
)

var GetGroupChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {

})
