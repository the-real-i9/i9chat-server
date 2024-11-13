package dmChatRoutes

import (
	dmcc "i9chat/controllers/chatControllers/dmChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/create_dm_chat_ack_msgs", dmcc.CreateNewDMChatAndAckMessages)
	router.Get("/chat_history", dmcc.GetChatHistory)
	router.Get("/:dm_chat_id/send_message", dmcc.SendMessage)
}
