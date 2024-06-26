package chatControllers

import (
	"fmt"
	"i9chat/services/appObservers"
	"i9chat/services/appServices"
	"i9chat/services/chatService"
	"i9chat/services/userService"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetDMChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		DmChatId int
		Offset   int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		// log.Println(r_err)
		return
	}

	dmChatHistory, app_err := chatService.DMChat{Id: body.DmChatId}.GetChatHistory(body.Offset)

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
	} else {
		w_err = c.WriteJSON(map[string]any{
			"statusCode": 200,
			"body":       dmChatHistory,
		})
	}

	if w_err != nil {
		// log.Println(w_err)
		return
	}
})

// this handler receives message acknowlegement for sent messages,
// and in turn changes the delivery status of messages sent by the child goroutine
var OpenDMMessagingStream = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var dmChatId int

	fmt.Sscanf(c.Params("dm_chat_id"), "%d", &dmChatId)

	var myMailbox = make(chan map[string]any, 3)

	mailboxKey := fmt.Sprintf("user-%d--dmchat-%d", user.UserId, dmChatId)

	dcso := appObservers.DMChatSessionObserver{}

	dcso.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		dcso.Unsubscribe(mailboxKey)
	}

	go sendDMChatMessages(c, user, dmChatId, endSession)

	/* ---- stream dm chat message events pending dispatch to the channel ---- */
	// observe that this happens once every new connection
	// A "What did I miss?" sort of query
	if eventDataList, err := (userService.User{Id: user.UserId}).GetDMChatMessageEventsPendingReceipt(dmChatId); err == nil {
		for _, eventData := range eventDataList {
			eventData := *eventData
			myMailbox <- eventData
		}
	}

	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			// log.Println(w_err)
			endSession()
			return
		}
	}
})

func sendDMChatMessages(c *websocket.Conn, user appTypes.JWTUserData, dmChatId int, endSession func()) {
	// this goroutine sends messages
	var body struct {
		Msg map[string]any
		At  time.Time
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			// log.Println(r_err)
			endSession()
			return
		}

		data, app_err := chatService.DMChat{Id: dmChatId}.SendMessage(
			user.UserId,
			appServices.MessageBinaryToUrl(user.UserId, body.Msg),
			body.At,
		)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(data)
		}

		if w_err != nil {
			// log.Println(w_err)
			endSession()
			return
		}
	}
}
