package dmChatRoutes

import (
	dmcc "i9chat/controllers/chatControllers/dmChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Post("/new", dmcc.CreateNewDMChat)
	router.Get("/chat_history", dmcc.GetChatHistory)
	router.Post("/:dm_chat_id/send_message", dmcc.SendMessage)
}
