package userroutes

import (
	"controlers/usercontrollers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/change_profile_picture", usercontrollers.ChangeProfilePicture)
	router.Get("/my_chats", usercontrollers.GetMyChats)
	router.Get("/all_users", usercontrollers.GetAllUsers)
	router.Get("/init_dm_chat_stream", usercontrollers.InitDMChatStream)
	router.Get("/init_group_chat_stream", usercontrollers.InitGroupChatStream)
}
