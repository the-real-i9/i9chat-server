package userRoutes

import (
	"i9chat/controllers/userControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/change_profile_picture", userControllers.ChangeProfilePicture)
	router.Get("/my_chats", userControllers.GetMyChats)
	router.Get("/all_users", userControllers.GetAllUsers)
	router.Get("/open_dm_chat_stream", userControllers.OpenDMChatStream)
	router.Get("/open_group_chat_stream", userControllers.OpenGroupChatStream)
}
