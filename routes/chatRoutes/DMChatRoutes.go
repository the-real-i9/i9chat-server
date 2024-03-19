package chatroutes

import (
	"controllers/chatcontrollers"

	"github.com/gofiber/fiber/v2"
)

func InitDMChat(router fiber.Router) {
	router.Get("/chat_history", chatcontrollers.GetDMChatHistory)
	router.Get("/watch_messsage", chatcontrollers.WatchDMChatMessage)
	router.Get("/send_message", chatcontrollers.SendDMChatMessage)
}
