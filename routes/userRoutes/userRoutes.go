package userroutes

import (
	"controlers/usercontrollers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/change_profile_picture", usercontrollers.ChangeProfilePicture)
	router.Get("/new_dm_chat", usercontrollers.NewDMChat)
	router.Get("/new_group_chat", usercontrollers.NewGroupChat)
	router.Get("/my_chats", usercontrollers.GetMyChats)
	router.Get("/watch_chat", usercontrollers.WatchChat)
}
