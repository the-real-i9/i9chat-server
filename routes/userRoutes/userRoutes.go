package userRoutes

import (
	UC "i9chat/controllers/userControllers"
	"i9chat/middlewares"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	am := middlewares.Auth

	router.Use("/go_online", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	router.Get("/go_online", am, UC.OpenWSStream)

	router.Post("/change_profile_picture", am, UC.ChangeProfilePicture)
	router.Post("/update_my_geolocation", am, UC.UpdateMyLocation)

	router.Get("/all_users", am, UC.GetAllUsers)
	router.Get("/search_user", am, UC.SearchUser)
	router.Get("/find_nearby_users", am, UC.FindNearbyUsers)

	router.Get("/my_chats", am, UC.GetMyChats)
}
