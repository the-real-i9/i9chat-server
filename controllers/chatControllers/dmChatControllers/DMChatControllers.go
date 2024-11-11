package dmChatControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	dmChat "i9chat/models/chatModel/dmChatModel"
	"i9chat/services/appServices"
	"i9chat/services/authServices"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetChatHistory = authServices.WSHandlerProtected(func(c *websocket.Conn) {

	var w_err error

	for {
		var body getChatHistoryBody

		if w_err != nil {
			log.Println(w_err)
			return
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		dmChatMessages, app_err := dmChat.GetChatHistory(body.DMChatId, body.Offset)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       dmChatMessages,
		})
	}
})

var SendMessage = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var dmChatId int

	_, err := fmt.Sscanf(c.Params("dm_chat_id"), "%d", &dmChatId)
	if err != nil {
		panic(err)
	}

	var w_err error

	for {
		var body sendMessageBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		senderData, app_err := sendMessage(
			dmChatId,
			clientUser.Id,
			appServices.UploadMessageMedia(clientUser.Id, body.Msg),
			body.At,
		)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(senderData)

	}
})
