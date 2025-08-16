package groupChatRoutes

import (
	GCC "i9chat/src/controllers/groupChatControllers"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
	router.Post("/new", GCC.CreateNewGroupChat)
	router.Get("/:group_id/membership_info", GCC.GetGroupMembershipInfo)
	router.Get("/:group_id/history", GCC.GetGroupChatHistory)
	router.Post("/:group_id/execute_action/:action", GCC.ExecuteAction)
}
