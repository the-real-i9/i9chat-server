package userroutes

import (
	"controlers/usercontrollers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/my_chats", usercontrollers.GetMyChats)
}
