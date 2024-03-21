package chatcontrollers

import (
	"fmt"
	"log"
	"services/chatservice"
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

func handleDMChatMessageSendingAndAcknowledgement(c *websocket.Conn, user apptypes.JWTUserData, dmChatId int, endSession func()) {
	var body struct {
		Action string
		Data   map[string]any
	}

	var msgBody struct {
		MsgContent map[string]any
		CreatedAt  time.Time
	}

	var ackBody struct {
		MsgId    int
		SenderId int
		Status   string
		At       time.Time
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			endSession()
			return
		}

		handleMessaging := func() error {
			helpers.MapToStruct(body.Data, &msgBody)

			data, app_err := chatservice.DMChat{Id: dmChatId}.SendMessage(user.UserId, msgBody.MsgContent, msgBody.CreatedAt)

			var w_err error
			if app_err != nil {
				w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
			} else {
				w_err = c.WriteJSON(data)
			}

			if w_err != nil {
				return w_err
			}

			return nil
		}

		/* The client application by itself sends acknowledgement, not the user.
		Each received message must be acknowledge as soon as it is received */
		handleAcknowledgement := func() {
			helpers.MapToStruct(body.Data, &ackBody)

			go chatservice.DMChatMessage{
				Id:       ackBody.MsgId,
				DmChatId: dmChatId,
				SenderId: ackBody.SenderId,
			}.UpdateDeliveryStatus(user.UserId, ackBody.Status, ackBody.At)

		}

		if body.Action == "messaging" {
			if err := handleMessaging(); err != nil {
				log.Println(err)
				endSession()
				return
			}
		} else if body.Action == "acknowledgement" {
			handleAcknowledgement()
		} else {
			w_err := c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, fmt.Errorf("invalid 'action' value. should be 'messaging' for sending messages, or 'acknowledgement' for acknowledging received messages")))
			if w_err != nil {
				endSession()
				return
			}
		}
	}
}

var InitDMChatSession = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	// This acts like a pipe handling send(), receive(), and acknowledgement()
	// New message and Message updates will be received in this session
	// Only New message is acknowledged.
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	c.WriteJSON(map[string]string{"msg": "First send the { dmChatId: #id } for this send() <-> receive() session."})

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

	dcmo := appglobals.DMChatMessageObserver{}

	dcmo.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		dcmo.Unsubscribe(mailboxKey)
	}

	go handleDMChatMessageSendingAndAcknowledgement(c, user, body.DmChatId, endSession)

	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			log.Println(w_err)
			endSession()
			return
		}
	}
})

/*
var WatchDMChatMessage = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	c.WriteJSON(map[string]string{"msg": "Enter the {dmChatId: {id}} to start watching. Any message after this closes the connection."})

	var body struct {
		DmChatId int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		log.Println(r_err)
		return
	}

	var myMailbox = make(chan map[string]any, 2)

	mailboxKey := fmt.Sprintf("user-%d--dmchat-%d", user.UserId, body.DmChatId)

	dcmo := appglobals.DMChatMessageObserver{}

	dcmo.Subscribe(mailboxKey, myMailbox)

	go func() {
		c.ReadMessage()
		dcmo.Unsubscribe(mailboxKey)
	}()

	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			log.Println(w_err)
			dcmo.Unsubscribe(mailboxKey)
			return
		}
	}
})

var SendDMChatMessage = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		DmChatId   int
		MsgContent map[string]any
		CreatedAt  time.Time
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		data, app_err := chatservice.DMChat{Id: body.DmChatId}.SendMessage(user.UserId, body.MsgContent, body.CreatedAt)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(data)
		}

		if w_err != nil {
			log.Println(w_err)
			return
		}
	}
})

var UpdateDMChatMessageDeliveryStatus = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	for {
		var body struct {
			MsgId    int
			DmChatId int
			SenderId int
			Status   string
			At       time.Time
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		go chatservice.DMChatMessage{
			Id:       body.MsgId,
			DmChatId: body.DmChatId,
			SenderId: body.SenderId,
		}.UpdateDeliveryStatus(user.UserId, body.Status, body.At)
	}
})
*/

var BatchUpdateDMChatMessageDeliveryStatus = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	for {
		var body struct {
			Status   string
			MsgDatas []*apptypes.DMChatMsgDeliveryData
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		go chatservice.BatchUpdateDMChatMessageDeliveryStatus(user.UserId, body.Status, body.MsgDatas)

	}

})
