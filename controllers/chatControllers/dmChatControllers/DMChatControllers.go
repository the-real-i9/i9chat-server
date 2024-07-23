package dmChatControllers

import (
	"fmt"
	dmChat "i9chat/models/chatModel/dmChatModel"
	user "i9chat/models/userModel"
	"i9chat/services/appObservers"
	"i9chat/services/appServices"
	"i9chat/services/chatService/dmChatService"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {

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

// this handler receives message acknowlegement for messages sent in an active chat,
// and in turn changes the delivery status of messages sent by the child goroutine
var OpenMessagingStream = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var dmChatId int

	_, param_err := fmt.Sscanf(c.Params("dm_chat_id"), "%d", &dmChatId)
	if param_err != nil {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusBadRequest, fmt.Errorf("parameter dm_chat_id is not a number"))); w_err != nil {
			log.Println(w_err)
		}
		return
	}

	var myMailbox = make(chan map[string]any, 3)

	mailboxKey := fmt.Sprintf("user-%d--dmchat-%d", clientUser.Id, dmChatId)

	dcso := appObservers.DMChatSessionObserver{}

	dcso.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		dcso.Unsubscribe(mailboxKey)
	}

	go sendMessages(c, clientUser, dmChatId, endSession)

	/* ---- stream dm chat message events pending dispatch to the channel ---- */
	// observe that this happens once every new connection
	// A "What did I miss?" sort of query
	if eventDataList, err := user.GetDMChatMessageEventsPendingReceipt(clientUser.Id, dmChatId); err == nil {
		for _, eventData := range eventDataList {
			eventData := *eventData
			myMailbox <- eventData
		}
	}

	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			log.Println(w_err)
			endSession()
			break
		}
	}
})

// this goroutine sends messages
func sendMessages(c *websocket.Conn, clientUser *appTypes.ClientUser, dmChatId int, endSession func()) {

	var w_err error

	for {
		var body openMessagingStreamBody

		if w_err != nil {
			log.Println(w_err)
			endSession()
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			endSession()
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		senderData, app_err := dmChatService.SendMessage(
			dmChatId,
			clientUser.Id,
			appServices.MessageBinaryToUrl(clientUser.Id, body.Msg),
			body.At,
		)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(senderData)

	}
}
