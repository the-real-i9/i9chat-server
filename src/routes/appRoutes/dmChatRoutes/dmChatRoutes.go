package dmChatRoutes

import (
	"i9chat/src/controllers/dmChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
	router.Get("/:partner_username/history", dmChatControllers.GetDMChatHistory)
}
