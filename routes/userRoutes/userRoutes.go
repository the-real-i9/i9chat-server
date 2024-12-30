package userRoutes

import (
	UC "i9chat/controllers/userControllers"
	"i9chat/middlewares"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	am := middlewares.Auth

	router.Get("/go_online", am, UC.OpenWSStream)

	router.Get("/change_profile_picture", am, UC.ChangeProfilePicture)
	router.Get("/update_my_geolocation", am, UC.UpdateMyLocation)

	router.Get("/all_users", am, UC.GetAllUsers)
	router.Get("/search_user", am, UC.SearchUser)
	router.Get("/find_nearby_users", am, UC.FindNearbyUsers)

	router.Get("/my_chats", am, UC.GetMyChats)
}
