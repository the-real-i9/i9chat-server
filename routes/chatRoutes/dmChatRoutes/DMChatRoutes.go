package dmChatRoutes

import (
	dcc "i9chat/controllers/chatControllers/dmChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/chat_history", dcc.GetChatHistory)
	router.Get("/open_messaging_stream/:dm_chat_id", dcc.OpenMessagingStream)
}
