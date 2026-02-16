package appRoutes

import (
	CUC "i9chat/src/controllers/chatControllers/chatUploadControllers"
	"i9chat/src/middlewares/authMiddlewares"
	"i9chat/src/routes/appRoutes/directChatRoutes"
	"i9chat/src/routes/appRoutes/groupChatRoutes"
	"i9chat/src/routes/appRoutes/realtimeRoute"
	"i9chat/src/routes/appRoutes/userRoutes"

	"github.com/gofiber/fiber/v3"
)

func Route(router fiber.Router) {
	router.Use(authMiddlewares.UserAuth)

	router.Route("/user", userRoutes.Route)

	router.Route("/ws", realtimeRoute.Route)

	router.Route("/group_chat", groupChatRoutes.Route)

	router.Route("/dm_chat", directChatRoutes.Route)

	router.Post("/chat_upload/authorize", CUC.AuthorizeUpload)
	router.Post("/chat_upload/authorize/visual", CUC.AuthorizeVisualUpload)
}
