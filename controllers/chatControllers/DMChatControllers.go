package chatcontrollers

import (
	"fmt"
	"log"
	"services/chatservice"
	"services/userservice"
	"time"
	"utils/appglobals"
	"utils/apptypes"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetDMChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		DmChatId int
		Offset   int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		log.Println(r_err)
		return
	}

	dmChatHistory, app_err := chatservice.DMChat{Id: body.DmChatId}.GetChatHistory(body.Offset)

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
	} else {
		w_err = c.WriteJSON(map[string]any{"chat_history": dmChatHistory})
	}

	if w_err != nil {
		log.Println(w_err)
		return
	}
})

var ActivateDMChatSession = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	// this goroutine receives message acknowlegement for sent messages
	// and in turn changes the delivery status of messages sent by the child goroutine
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	c.WriteJSON(map[string]string{"msg": "First send the { dmChatId: #id } for this send(), ack() <-> recv_ack() session."})

	var body struct {
		DmChatId int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		log.Println(r_err)
		return
	}

	var myMailbox = make(chan map[string]any, 3)

	mailboxKey := fmt.Sprintf("user-%d--dmchat-%d", user.UserId, body.DmChatId)

	dcmo := appglobals.DMChatSessionObserver{}

	dcmo.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		dcmo.Unsubscribe(mailboxKey)
	}

	go sendDMChatMessages(c, user, body.DmChatId, endSession)

	/* ---- stream dm chat message events pending dispatch to the channel ---- */
	// observe that this happens once every new connection
	// A "What did I miss?" sort of query
	if event_data_kvps, err := (userservice.User{Id: user.UserId}).GetDMChatMessageEventsPendingDispatch(body.DmChatId); err == nil {
		for _, evk := range event_data_kvps {
			evk := *evk
			myMailbox <- evk
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

func sendDMChatMessages(c *websocket.Conn, user apptypes.JWTUserData, dmChatId int, endSession func()) {
	// this goroutine sends messages
	var body struct {
		MsgContent map[string]any
		CreatedAt  time.Time
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			endSession()
			return
		}

		data, app_err := chatservice.DMChat{Id: dmChatId}.SendMessage(user.UserId, body.MsgContent, body.CreatedAt)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(data)
		}

		if w_err != nil {
			log.Println(w_err)
			endSession()
			return
		}
	}
}
