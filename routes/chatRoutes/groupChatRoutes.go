package chatroutes

import (
	"controllers/chatcontrollers"

	"github.com/gofiber/fiber/v2"
)

func InitGroupChat(router fiber.Router) {
	router.Get("/chat_history", chatcontrollers.GetGroupChatHistory)
	router.Get("/new_messsage", chatcontrollers.ListenForNewGroupChatMessage)
	router.Get("/send_message")
}
