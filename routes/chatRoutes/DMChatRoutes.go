package chatroutes

import (
	"controllers/chatcontrollers"

	"github.com/gofiber/fiber/v2"
)

func InitDMChat(router fiber.Router) {
	router.Get("/chat_history", chatcontrollers.GetDMChatHistory)
	// router.Get("/watch_message", chatcontrollers.WatchDMChatMessage)
	// router.Get("/send_message", chatcontrollers.SendDMChatMessage)
	// router.Get("/update_message_delivery_status", chatcontrollers.UpdateDMChatMessageDeliveryStatus)
	router.Get("/init_chat_session", chatcontrollers.InitDMChatSession)
	router.Get("/batch_update_message_delivery_status", chatcontrollers.BatchUpdateDMChatMessageDeliveryStatus)
}
