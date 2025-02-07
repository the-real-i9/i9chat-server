package dmChatRoutes

import (
	dmcc "i9chat/controllers/chatControllers/dmChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/:partner_username/history", dmcc.GetChatHistory)
}
