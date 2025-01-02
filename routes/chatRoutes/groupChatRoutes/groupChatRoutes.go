package groupChatRoutes

import (
	grcc "i9chat/controllers/chatControllers/groupChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Post("/create_group_chat", grcc.CreateNewGroupChat)
	router.Get("/chat_history", grcc.GetChatHistory)
	router.Post("/:group_chat_id/send_message", grcc.SendMessage)
	router.Post("/execute_action", grcc.ExecuteAction)
}
