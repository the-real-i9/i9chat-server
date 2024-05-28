package chatRoutes

import (
	"i9chat/controllers/chatControllers"

	"github.com/gofiber/fiber/v2"
)

func InitGroupChat(router fiber.Router) {
	router.Get("/chat_history", chatControllers.GetGroupChatHistory)
	router.Get("/open_dm_messaging_stream/:group_chat_id", chatControllers.OpenGroupMessagingStream)
}
