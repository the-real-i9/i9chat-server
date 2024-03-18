package usercontrollers

import (
	"fmt"
	"log"
	"services/userservice"
	"utils/appglobals"
	"utils/apptypes"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
)

var GetMyChats = helpers.WSHandlerProtected(func(c *websocket.Conn) {

	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	myChats, app_err := userservice.GetMyChats(user.UserId)

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.AppError(500, app_err))
	} else {
		w_err = c.WriteJSON(map[string]any{"my_chats": myChats})
	}

	if w_err != nil {
		log.Println(w_err)
		return
	}

})

var ListenForNewChat = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	// a data channel for transmitting data
	var myMailbox = make(chan map[string]any, 5)
	// a control channel for terminating the wait
	var closeMailBox = make(chan int)

	mailboxKey := fmt.Sprintf("user-%d", user.UserId)

	// subscribe to receiving chat updates
	// myMailbox is passed by reference to an observer keeping several mailboxes wanting to receive updates
	nco := appglobals.NewChatObserver{}

	nco.Subscribe(mailboxKey, myMailbox)

	go func() {
		// a strategy to close the mailbox and, in turn, the websocket connection
		// Ideally, this route doesn't receive any message,
		// therefore, it'll be unnecessary to check for a specific "close" command
		// so any message received at all triggers a close
		c.ReadMessage()
		closeMailBox <- 1
		close(closeMailBox)
	}()

	for {
		// a data channel and a control channel is watched by the select
		select {
		case data := <-myMailbox:
			w_err := c.WriteJSON(map[string]any{"new_chat": data})
			if w_err != nil {
				log.Println(w_err)
				nco.Unsubscribe(mailboxKey)
				return
			}
		case <-closeMailBox:
			nco.Unsubscribe(mailboxKey)
			return
		}
	}
})
