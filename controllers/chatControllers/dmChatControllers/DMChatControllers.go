package dmChatControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/services/chatServices/dmChatService"
	"i9chat/services/utils/authUtilServices"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var CreateNewDMChatAndAckMessages = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var w_err error

	for {
		var body createNewDMChatAndAckMessagesBody

		var newChatData newDMChatDataT

		// For DM Chat, we allowed options for both single and batch acknowledgements
		var ackMsgData ackMsgDataT

		var batchAckMsgData batchAckMsgDataT

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

		createNewChat := func() error {

			helpers.MapToStruct(body.Data, &newChatData)

			if val_err := newChatData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			respData, app_err := dmChatService.NewDMChat(
				clientUser.Id,
				newChatData.PartnerId,
				newChatData.InitMsg,
				newChatData.CreatedAt,
			)

			if app_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			}

			return c.WriteJSON(respData)

		}

		// acknowledge messages singly
		acknowledgeMessage := func() error {

			helpers.MapToStruct(body.Data, &ackMsgData)

			if val_err := ackMsgData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			go dmChatService.UpdateMessageDeliveryStatus(ackMsgData.DMChatId, ackMsgData.MsgId, ackMsgData.SenderId, clientUser.Id, ackMsgData.Status, ackMsgData.At)

			return nil
		}

		// acknowledge messages in batch
		batchAcknowledgeMessages := func() error {

			helpers.MapToStruct(body.Data, &batchAckMsgData)

			if val_err := batchAckMsgData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			go dmChatService.BatchUpdateMessageDeliveryStatus(clientUser.Id, batchAckMsgData.Status, batchAckMsgData.MsgAckDatas)

			return nil
		}

		if body.Action == "create new chat" {

			w_err = createNewChat()

		} else if body.Action == "acknowledge message" {

			w_err = acknowledgeMessage()

		} else if body.Action == "batch acknowledge messages" {

			w_err = batchAcknowledgeMessages()

		} else {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("invalid 'action' value")))
		}
	}
})

var GetChatHistory = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {

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

		respData, app_err := dmChatService.GetChatHistory(body.DMChatId, body.Offset)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       respData,
		})
	}
})

var SendMessage = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
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

		respData, app_err := dmChatService.SendMessage(
			dmChatId,
			clientUser.Id,
			body.Msg,
			body.At,
		)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(respData)

	}
})
