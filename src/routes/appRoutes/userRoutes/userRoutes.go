package userRoutes

import (
	UC "i9chat/src/controllers/userControllers"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {

	router.Get("/session_user", UC.GetSessionUser)
	router.Get("/my_profile", UC.GetMyProfile)

	router.Post("/change_profile_picture", UC.ChangeProfilePicture)
	router.Post("/set_geolocation", UC.SetMyLocation)

	router.Get("/find_user", UC.FindUser)
	router.Get("/find_nearby_users", UC.FindNearbyUsers)

	router.Get("/my_chats", UC.GetMyChats)

	router.Get("/signout", UC.SignOut)
}
