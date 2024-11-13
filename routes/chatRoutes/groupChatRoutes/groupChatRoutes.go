package groupChatRoutes

import (
	grcc "i9chat/controllers/chatControllers/groupChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/create_group_chat_ack_msgs", grcc.CreateNewGroupChatAndAckMessages)
	router.Get("/chat_history", grcc.GetChatHistory)
	router.Get("/:group_chat_id/send_message", grcc.SendMessage)
	router.Get("/execute_action", grcc.ExecuteAction)
}
