package userRoutes

import (
	UC "i9chat/src/controllers/userControllers"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Use("/go_online", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	router.Get("/go_online", UC.OpenWSStream)

	router.Get("/my_profile", UC.GetMyProfile)

	router.Post("/change_profile_picture", UC.ChangeProfilePicture)
	router.Post("/change_phone_number", UC.ChangePhone)
	router.Post("/update_geolocation", UC.UpdateMyLocation)

	router.Get("/find_user", UC.FindUser)
	router.Get("/find_nearby_users", UC.FindNearbyUsers)

	router.Get("/my_chats", UC.GetMyChats)

	router.Get("/signout", UC.SignOut)
}
