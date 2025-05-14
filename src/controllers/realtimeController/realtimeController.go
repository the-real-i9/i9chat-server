package realtimeController

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/eventStreamService"
	"i9chat/src/services/userService"
	"log"

	"github.com/gofiber/contrib/websocket"
)

var WSStream = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	go userService.GoOnline(context.Background(), clientUser.Username)

	eventStreamService.Subscribe(clientUser.Username, c)

	var w_err error

	for {
		var body clientEventBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.WSErrReply(val_err, body.Event))
			continue
		}

		switch body.Event {
		case "send dm chat message":
			respData, err := sendDMChatMsgHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}
			c.WriteJSON(helpers.WSReply(respData, body.Event))
		case "ack dm chat message delivered":

			err := ackDMChatMsgDeliveredHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			c.WriteJSON(helpers.WSReply(true, body.Event))
		case "ack dm chat message read":

			err := ackDMChatMsgReadHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			c.WriteJSON(helpers.WSReply(true, body.Event))
		case "get dm chat history":

			respData, err := getDMChatHistoryHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			c.WriteJSON(helpers.WSReply(respData, body.Event))
		case "send group chat message":

			respData, err := sendGroupChatMsgHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			c.WriteJSON(helpers.WSReply(respData, body.Event))
		case "ack group chat message delivered":

			respData, err := ackGroupChatMsgDeliveredHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			c.WriteJSON(helpers.WSReply(respData, body.Event))
		case "ack group chat message read":

			respData, err := ackGroupChatMsgReadHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			c.WriteJSON(helpers.WSReply(respData, body.Event))
		case "get group chat history":

			respData, err := getGroupChatHistoryHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			c.WriteJSON(helpers.WSReply(respData, body.Event))
		case "get group info":

			respData, err := getGroupInfoHndl(ctx, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrReply(err, body.Event))
				continue
			}

			c.WriteJSON(helpers.WSReply(respData, body.Event))

		default:
			w_err = c.WriteJSON(helpers.WSErrReply(fmt.Errorf("invalid event: %s", body.Event), body.Event))
			continue
		}
	}

	go userService.GoOffline(context.Background(), clientUser.Username)

	eventStreamService.Unsubscribe(clientUser.Username)
})
