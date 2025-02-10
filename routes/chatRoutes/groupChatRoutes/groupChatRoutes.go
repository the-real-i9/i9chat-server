package groupChatRoutes

import (
	grcc "i9chat/controllers/chatControllers/groupChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Post("/new", grcc.CreateNewGroupChat)
	router.Get("/:group_id/chat_history", grcc.GetChatHistory)
	router.Post("/:group_id/execute_action/:action", grcc.ExecuteAction)
}
