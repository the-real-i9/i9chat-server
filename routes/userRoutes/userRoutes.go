package userRoutes

import (
	UC "i9chat/controllers/userControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/go_online", UC.GoOnline)

	router.Get("/change_profile_picture", UC.ChangeProfilePicture)
	router.Get("/update_my_geolocation", UC.UpdateMyLocation)

	router.Get("/all_users", UC.GetAllUsers)
	router.Get("/search_user", UC.SearchUser)
	router.Get("/find_nearby_users", UC.FindNearbyUsers)

	router.Get("/my_chats", UC.GetMyChats)
}
