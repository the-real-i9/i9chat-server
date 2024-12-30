package groupChatRoutes

import (
	grcc "i9chat/controllers/chatControllers/groupChatControllers"
	"i9chat/middlewares"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	am := middlewares.Auth

	router.Post("/create_group_chat", am, grcc.CreateNewGroupChat)
	router.Get("/chat_history", am, grcc.GetChatHistory)
	router.Post("/:group_chat_id/send_message", am, grcc.SendMessage)
	router.Post("/execute_action", am, grcc.ExecuteAction)
}
