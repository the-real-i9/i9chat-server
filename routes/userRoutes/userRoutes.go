package userRoutes

import (
	UC "i9chat/controllers/userControllers"

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

	router.Post("/change_profile_picture", UC.ChangeProfilePicture)
	router.Post("/update_my_geolocation", UC.UpdateMyLocation)

	router.Get("/search_user", UC.SearchUser)
	router.Get("/find_nearby_users", UC.FindNearbyUsers)

	router.Get("/my_chats", UC.GetMyChats)

	router.Get("/logout", UC.Logout)
}
