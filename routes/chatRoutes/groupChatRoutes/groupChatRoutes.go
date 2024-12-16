package groupChatRoutes

import (
	grcc "i9chat/controllers/chatControllers/groupChatControllers"
	"i9chat/middlewares"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	am := middlewares.Auth

	router.Get("/create_group_chat_ack_msgs", am, grcc.CreateNewGroupChatAndAckMessages)
	router.Get("/chat_history", am, grcc.GetChatHistory)
	router.Get("/:group_chat_id/send_message", am, grcc.SendMessage)
	router.Get("/execute_action", am, grcc.ExecuteAction)
}
