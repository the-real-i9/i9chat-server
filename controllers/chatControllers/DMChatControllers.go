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
