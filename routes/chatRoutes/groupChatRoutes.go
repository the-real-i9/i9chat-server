package chatroutes

import (
	"controllers/chatcontrollers"

	"github.com/gofiber/fiber/v2"
)

func InitGroupChat(router fiber.Router) {
	router.Get("/chat_history", chatcontrollers.GetGroupChatHistory)
	router.Get("/watch_message", chatcontrollers.WatchGroupChatMessage)
	router.Get("/send_message", chatcontrollers.SendGroupChatMessage)
	router.Get("/watch_activity", chatcontrollers.WatchGroupActivity)
}
