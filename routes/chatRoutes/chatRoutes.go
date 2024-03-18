package chatroutes

import (
	"controllers/chatcontrollers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/new_messsage", chatcontrollers.ListenForNewMessage)
}
