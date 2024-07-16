package userRoutes

import (
	UC "i9chat/controllers/userControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/change_profile_picture", UC.ChangeProfilePicture)
	router.Get("/my_chats", UC.GetMyChats)
	router.Get("/all_users", UC.GetAllUsers)
	router.Get("/search_user", UC.SearchUser)
	router.Get("/find_nearby_users", UC.FindNearbyUsers)
	router.Get("/update_my_geolocation", UC.UpdateMyGeolocation)
	router.Get("/switch_my_presence", UC.SwitchMyPresence)
	router.Get("/open_dm_chat_stream", UC.OpenDMChatStream)
	router.Get("/open_group_chat_stream", UC.OpenGroupChatStream)
}
