package chatcontrollers

import (
	"fmt"
	"log"
	"services/chatservice"
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

	dmChat := chatservice.DMChat{Id: body.DmChatId}

	dmChatHistory, app_err := dmChat.GetChatHistory(body.Offset)

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

var ListenForNewDMChatMessage = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		ChatId int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		log.Println(r_err)
		return
	}

	var myMailbox = make(chan map[string]any, 2)
	var closeMailBox = make(chan int)

	mailboxKey := fmt.Sprintf("user-%d--dmchat-%d", user.UserId, body.ChatId)

	appglobals.SubscribeToNewDMChatMessageUpdate(mailboxKey, myMailbox)

	go func() {
		c.ReadMessage()
		closeMailBox <- 1
		close(closeMailBox)
	}()

	for {
		select {
		case data := <-myMailbox:
			w_err := c.WriteJSON(map[string]any{"new_message": data})
			if w_err != nil {
				log.Println(w_err)
				appglobals.UnsubscribeFromNewDMChatMessageUpdate(mailboxKey)
				return
			}
		case <-closeMailBox:
			appglobals.UnsubscribeFromNewDMChatMessageUpdate(mailboxKey)
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
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		dmChat := chatservice.DMChat{Id: body.DmChatId}
		data, app_err := dmChat.SendMessage(user.UserId, body.MsgContent)

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
