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
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {

	var body struct {
		DMChatId int
		Offset   int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		log.Println(r_err)
		return
	}

	dmChatMessages, app_err := dmChat.GetChatHistory(body.DMChatId, body.Offset)

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
	} else {
		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       dmChatMessages,
		})
	}

	if w_err != nil {
		log.Println(w_err)
		return
	}
})

// this handler receives message acknowlegement for sent messages,
// and in turn changes the delivery status of messages sent by the child goroutine
var OpenMessagingStream = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var dmChatId int

	fmt.Sscanf(c.Params("dm_chat_id"), "%d", &dmChatId)

	var myMailbox = make(chan map[string]any, 3)

	mailboxKey := fmt.Sprintf("user-%d--dmchat-%d", clientUser.Id, dmChatId)

	dcso := appObservers.DMChatSessionObserver{}

	dcso.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		dcso.Unsubscribe(mailboxKey)
	}

	go sendDMChatMessages(c, clientUser, dmChatId, endSession)

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
			return
		}
	}
})

// this goroutine sends messages
func sendDMChatMessages(c *websocket.Conn, clientUser *appTypes.ClientUser, dmChatId int, endSession func()) {
	var body struct {
		Msg map[string]any
		At  time.Time
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			endSession()
			return
		}

		senderData, app_err := dmChatService.SendMessage(
			dmChatId,
			clientUser.Id,
			appServices.MessageBinaryToUrl(clientUser.Id, body.Msg),
			body.At,
		)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(senderData)
		}

		if w_err != nil {
			log.Println(w_err)
			endSession()
			return
		}
	}
}
