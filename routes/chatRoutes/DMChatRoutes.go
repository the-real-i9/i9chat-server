package chatroutes

import (
	"controllers/chatcontrollers"

	"github.com/gofiber/fiber/v2"
)

func InitDMChat(router fiber.Router) {
	router.Get("/chat_history", chatcontrollers.GetDMChatHistory)
}
