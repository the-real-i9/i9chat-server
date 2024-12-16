package dmChatRoutes

import (
	dmcc "i9chat/controllers/chatControllers/dmChatControllers"
	"i9chat/middlewares"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	am := middlewares.Auth

	router.Get("/create_dm_chat_ack_msgs", am, dmcc.CreateNewDMChatAndAckMessages)
	router.Get("/chat_history", am, dmcc.GetChatHistory)
	router.Get("/:dm_chat_id/send_message", am, dmcc.SendMessage)
}
