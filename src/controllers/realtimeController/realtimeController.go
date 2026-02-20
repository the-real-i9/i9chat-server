package realtimeController

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/controllers/chatControllers/directChatControllers"
	"i9chat/src/controllers/chatControllers/groupChatControllers"
	"i9chat/src/helpers"
	"i9chat/src/services/realtimeService"
	"i9chat/src/services/userService"
	"log"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

var WSStream = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	go userService.GoOnline(context.Background(), clientUser.Username)

	realtimeService.AddPipe(ctx, clientUser.Username, c)

	var w_err error

	for {

		if w_err != nil {
			log.Println(w_err)
			break
		}

		MSG_TYPE, msgPackBt, r_err := c.ReadMessage()
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if MSG_TYPE != websocket.BinaryMessage {
			w_err = c.WriteMessage(websocket.BinaryMessage, helpers.ToBtMsgPack(fmt.Errorf("unexpected message type")))
			continue
		}

		body := helpers.FromBtMsgPack[rtActionBody](msgPackBt)

		if err := body.Validate(); err != nil {
			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
			continue
		}

		cancelUserPresenceSub := make(map[string]context.CancelFunc)

		switch body.Action {
		case "subscribe to user presence change":

			data := helpers.FromBtMsgPack[subToUserPresenceAcd](body.Data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			for _, tu := range data.Usernames {
				ctx, cancel := context.WithCancel(ctx)

				realtimeService.SubscribeToUserPresence(ctx, clientUser.Username, tu, cancel)

				cancelUserPresenceSub[tu] = cancel
			}
		case "unsubscribe from user presence change":

			data := helpers.FromBtMsgPack[unsubFromUserPresenceAcd](body.Data)

			if err := data.Validate(); err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			for _, tu := range data.Usernames {
				if cancel, ok := cancelUserPresenceSub[tu]; ok {
					cancel()
				}

				delete(cancelUserPresenceSub, tu)
			}
		case "direct chat: send message":
			respData, err := directChatControllers.SendMessage(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(respData, body.Action)))
		case "direct chat: ack messages delivered":

			respData, err := directChatControllers.AckMessagesDelivered(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(respData, body.Action)))
		case "direct chat: ack messages read":

			respData, err := directChatControllers.AckMessagesRead(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(respData, body.Action)))
		case "group chat: send message":

			respData, err := groupChatControllers.SendMessage(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(respData, body.Action)))
		case "group chat: ack messages delivered":

			respData, err := groupChatControllers.AckMessagesDelivered(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(respData, body.Action)))
		case "group chat: ack messages read":

			respData, err := groupChatControllers.AckMessagesRead(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(respData, body.Action)))
		case "group: get info":

			respData, err := groupChatControllers.GetGroupInfo(ctx, body.Data)
			if err != nil {
				w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(err, body.Action)))
				continue
			}

			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSReply(respData, body.Action)))

		default:
			w_err = c.WriteMessage(MSG_TYPE, helpers.ToBtMsgPack(helpers.WSErrReply(fiber.NewErrorf(fiber.StatusInternalServerError, "invalid event: %s", body.Action), body.Action)))
			continue
		}
	}

	go userService.GoOffline(context.Background(), clientUser.Username)

	realtimeService.RemovePipe(clientUser.Username)
})
