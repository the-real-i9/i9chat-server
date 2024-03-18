package chatcontrollers

import (
	"fmt"
	"log"
	"utils/appglobals"
	"utils/apptypes"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
)

var ListenForNewMessage = helpers.WSHandlerProtected(func(c *websocket.Conn) {
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

	mailboxKey := fmt.Sprintf("user-%d--chat-%d", user.UserId, body.ChatId)

	appglobals.SubscribeToNewMessageUpdate(mailboxKey, myMailbox)

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
				appglobals.UnsubscribeFromNewMessageUpdate(mailboxKey)
				return
			}
		case <-closeMailBox:
			appglobals.UnsubscribeFromNewMessageUpdate(mailboxKey)
			return
		}
	}
})
