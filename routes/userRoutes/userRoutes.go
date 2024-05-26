package userRoutes

import (
	"i9chat/controllers/userControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/change_profile_picture", userControllers.ChangeProfilePicture)
	router.Get("/my_chats", userControllers.GetMyChats)
	router.Get("/all_users", userControllers.GetAllUsers)
	router.Get("/search_user", userControllers.SearchUser)
	router.Get("/find_nearby_users", userControllers.FindNearbyUsers)
	router.Get("/update_my_location", userControllers.UpdateMyLocation)
	router.Get("/switch_my_presence", userControllers.SwitchMyPresence)
	router.Get("/open_dm_chat_stream", userControllers.OpenDMChatStream)
	router.Get("/open_group_chat_stream", userControllers.OpenGroupChatStream)
}
