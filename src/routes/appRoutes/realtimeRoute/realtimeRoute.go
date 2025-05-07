package realtimeRoute

import (
	RC "i9chat/src/controllers/realtimeController"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
	router.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	router.Get("", RC.WSStream)

}
