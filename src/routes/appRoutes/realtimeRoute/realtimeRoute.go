package realtimeRoute

import (
	RC "i9chat/src/controllers/realtimeController"

	"github.com/gofiber/fiber/v3"
)

func Route(router fiber.Router) {
	router.Use(func(c fiber.Ctx) error {
		if c.IsWebSocket() {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	router.Get("", RC.WSStream)

}
