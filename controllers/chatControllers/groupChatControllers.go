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

var GetGroupChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		GroupChatId int
		Offset      int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		log.Println(r_err)
		return
	}

	groupChatHistory, app_err := chatservice.GroupChat{Id: body.GroupChatId}.GetChatHistory(body.Offset)

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
	} else {
		w_err = c.WriteJSON(map[string]any{"chat_history": groupChatHistory})
	}

	if w_err != nil {
		log.Println(w_err)
		return
	}
})

var ListenForNewGroupChatMessage = helpers.WSHandlerProtected(func(c *websocket.Conn) {
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

	mailboxKey := fmt.Sprintf("user-%d--groupchat-%d", user.UserId, body.ChatId)

	ngcmo := appglobals.NewGroupChatMessageObserver{}

	ngcmo.Subscribe(mailboxKey, myMailbox)

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
				ngcmo.Unubscribe(mailboxKey)
				return
			}
		case <-closeMailBox:
			ngcmo.Unubscribe(mailboxKey)
			return
		}
	}
})

var SendGroupChatMessage = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		GroupChatId int
		MsgContent  map[string]any
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		groupChat := chatservice.GroupChat{Id: body.GroupChatId}
		data, app_err := groupChat.SendMessage(user.UserId, body.MsgContent)

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
