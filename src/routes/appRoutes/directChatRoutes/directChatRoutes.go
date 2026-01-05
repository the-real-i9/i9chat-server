package directChatRoutes

import (
	"i9chat/src/controllers/chatControllers/directChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
	router.Get("/:partner_username/history", directChatControllers.GetDirectChatHistory)
}
