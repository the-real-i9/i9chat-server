package chatRoutes

import (
	"i9chat/controllers/chatControllers"

	"github.com/gofiber/fiber/v2"
)

func InitDMChat(router fiber.Router) {
	router.Get("/chat_history", chatControllers.GetDMChatHistory)
	router.Get("/activate_chat_session", chatControllers.ActivateDMChatSession)
}
