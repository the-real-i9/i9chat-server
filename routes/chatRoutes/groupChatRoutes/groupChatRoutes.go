package groupChatRoutes

import (
	gccs "i9chat/controllers/chatControllers/groupChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/chat_history", gccs.GetChatHistory)
	router.Get("/open_messaging_stream/:group_chat_id", gccs.OpenMessagingStream)
	router.Get("/execute_action", gccs.ExecuteAction)
}
