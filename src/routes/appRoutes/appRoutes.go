package appRoutes

import (
	"i9chat/src/middlewares/authMiddlewares"
	"i9chat/src/routes/appRoutes/groupChatRoutes"
	"i9chat/src/routes/appRoutes/realtimeRoute"
	"i9chat/src/routes/appRoutes/userRoutes"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
	router.Use(authMiddlewares.UserAuth)

	router.Route("/user", userRoutes.Route)

	router.Route("/ws", realtimeRoute.Route)

	router.Route("/group_chat", groupChatRoutes.Route)
}
