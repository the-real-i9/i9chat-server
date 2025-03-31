package groupChatRoutes

import (
	grcc "i9chat/src/controllers/chatControllers/groupChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Post("/new", grcc.CreateNewGroupChat)
	router.Post("/:group_id/execute_action/:action", grcc.ExecuteAction)
}
