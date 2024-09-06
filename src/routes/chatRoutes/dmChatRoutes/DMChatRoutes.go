package dmChatRoutes

import (
	dcc "i9chat/src/controllers/chatControllers/dmChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/chat_history", dcc.GetChatHistory)
	router.Get("/:dm_chat_id/open_messaging_stream", dcc.OpenMessagingStream)
}
