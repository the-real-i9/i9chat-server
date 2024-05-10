package userRoutes

import (
	"i9chat/controllers/userControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/change_profile_picture", userControllers.ChangeProfilePicture)
	router.Get("/my_chats", userControllers.GetMyChats)
	router.Get("/all_users", userControllers.GetAllUsers)
	router.Get("/init_dm_chat_stream", userControllers.InitDMChatStream)
	router.Get("/init_group_chat_stream", userControllers.InitGroupChatStream)
}
