package groupChatRoutes

import (
	gccs "i9chat/src/controllers/chatControllers/groupChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/chat_history", gccs.GetChatHistory)
	router.Get("/:group_chat_id/open_messaging_stream", gccs.OpenMessagingStream)
	router.Get("/execute_action", gccs.ExecuteAction)
}
