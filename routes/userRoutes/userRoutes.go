package userroutes

import (
	"controlers/usercontrollers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/change_profile_picture")
	router.Get("/my_chats", usercontrollers.GetMyChats)
	router.Get("/watch_chat", usercontrollers.WatchChat)
}
