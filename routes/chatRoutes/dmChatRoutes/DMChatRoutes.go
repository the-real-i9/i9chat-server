package dmChatRoutes

import (
	dmcc "i9chat/controllers/chatControllers/dmChatControllers"
	"i9chat/middlewares"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	am := middlewares.Auth

	router.Post("/create_dm_chat", am, dmcc.CreateNewDMChat)
	router.Get("/chat_history", am, dmcc.GetChatHistory)
	router.Post("/:dm_chat_id/send_message", am, dmcc.SendMessage)
}
