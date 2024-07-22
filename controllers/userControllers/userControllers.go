package userControllers

import (
	"fmt"
	user "i9chat/models/userModel"
	"i9chat/services/appObservers"
	"i9chat/services/appServices"
	"i9chat/services/chatService/dmChatService"
	"i9chat/services/chatService/groupChatService"
	"i9chat/services/userService"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var ChangeProfilePicture = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var body changeProfilePictureBody

	var w_err error

	for {
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
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		if app_err := userService.ChangeMyProfilePicture(clientUser.Id, body.PictureData); app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("operation failed: %s", app_err)))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg": "Operation Successful",
			},
		})

	}
})

// This handler:
//
// 1. As soon as connection is restored (client online), streams all new dm chats pending receipt (while client offline) to the client, and keeps the connection open to send new ones.
//
// 2. Lets the client: "initiate a new dm chat" and "acknowledge received dm messages"
var OpenDMChatStream = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	// a channel for streaming data to client
	var myMailbox = make(chan map[string]any, 5)

	// subscribe to receiving chat updates
	// myMailbox is passed by reference to an observer keeping several mailboxes wanting to receive updates
	dco := appObservers.DMChatObserver{}

	mailboxKey := fmt.Sprintf("user-%d", clientUser.Id)

	dco.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		dco.Unsubscribe(mailboxKey)
	}

	go createNewDMChatAndAckMessages(c, clientUser, endSession)

	/* ---- stream chat events pending dispatch to the myMailbox ---- */
	// this excecutes once every connection restoration
	// "What did I miss?" - When the client comes online they get a report of all missed data (while offline)
	if eventDataList, err := user.GetDMChatEventsPendingReceipt(clientUser.Id); err == nil {
		for _, eventData := range eventDataList {
			eventData := *eventData
			myMailbox <- eventData
		}
	}

	// when myMailbox receives any new data, it is sent to the client
	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			log.Println(w_err)
			endSession()
			break
		}
	}
})

// This goroutine handles:
//
// + initating new dm chats
//
// + sending acknowledgements for received dm messages
func createNewDMChatAndAckMessages(c *websocket.Conn, clientUser *appTypes.ClientUser, endSession func()) {
	var body openDMChatStreamBody

	var newChatBody newDMChatBodyT

	// For DM Chat, we allowed options for both single and batch acknowledgements
	var ackMsgBody ackMsgBodyT

	var batchAckMsgBody batchAckMsgBodyT

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			endSession()
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			endSession()
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		createNewChat := func() error {

			helpers.MapToStruct(body.Data, &newChatBody)

			if val_err := newChatBody.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			initiatorData, app_err := dmChatService.NewDMChat(
				clientUser.Id,
				newChatBody.PartnerId,
				appServices.MessageBinaryToUrl(clientUser.Id, newChatBody.InitMsg),
				newChatBody.CreatedAt,
			)

			if app_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			}

			return c.WriteJSON(initiatorData)

		}

		// acknowledge messages singly
		acknowledgeMessage := func() error {

			helpers.MapToStruct(body.Data, &ackMsgBody)

			if val_err := ackMsgBody.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			go dmChatService.UpdateMessageDeliveryStatus(ackMsgBody.DMChatId, ackMsgBody.MsgId, ackMsgBody.SenderId, clientUser.Id, ackMsgBody.Status, ackMsgBody.At)

			return nil
		}

		// acknowledge messages in batch
		batchAcknowledgeMessages := func() error {

			helpers.MapToStruct(body.Data, &batchAckMsgBody)

			if val_err := batchAckMsgBody.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			go dmChatService.BatchUpdateMessageDeliveryStatus(clientUser.Id, batchAckMsgBody.Status, batchAckMsgBody.MsgAckDatas)

			return nil
		}

		if body.Action == "create new chat" {

			w_err = createNewChat()

		} else if body.Action == "acknowledge message" {

			w_err = acknowledgeMessage()

		} else if body.Action == "batch acknowledge messages" {

			w_err = batchAcknowledgeMessages()

		} else {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("invalid 'action' value")))
		}
	}
}

// This handler:
//
// 1. As soon as connection is restored (client online), streams all new group chats pending receipt (while client offline) to the client, and keeps the connection open to send new ones.
//
// 2. Lets the client: "initiate a new group chat" and "acknowledge received group messages"
var OpenGroupChatStream = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	// a channel for streaming data to client
	var myMailbox = make(chan map[string]any, 5)

	// subscribe to receiving chat updates
	// myMailbox is passed by reference to an observer keeping several mailboxes wanting to receive updates
	gco := appObservers.GroupChatObserver{}

	mailboxKey := fmt.Sprintf("user-%d", clientUser.Id)

	gco.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		gco.Unsubscribe(mailboxKey)
	}

	go createNewGroupChatAndAckMessages(c, clientUser, endSession)

	/* ---- stream chat events pending dispatch to the myMailbox ---- */
	// this excecutes once every connection restoration
	// "What did I miss?" - When the client comes online they get a report of all missed data (while offline)
	if eventDataList, err := user.GetGroupChatEventsPendingReceipt(clientUser.Id); err == nil {
		for _, eventData := range eventDataList {
			eventData := *eventData
			myMailbox <- eventData
		}
	}

	// when myMailbox receives any new data, it is sent to the client
	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			log.Println(w_err)
			endSession()
			return
		}
	}
})

// This goroutine handles:
//
// + initating new group chats
//
// + sending acknowledgement for received group messages
func createNewGroupChatAndAckMessages(c *websocket.Conn, clientUser *appTypes.ClientUser, endSession func()) {
	var body openGroupChatStreamBody

	var newChatBody newGroupChatBodyT

	// For Group chat, messages should be acknowledged in batches,
	// and it's only for a single group chat at a time
	var ackMsgsBody ackMsgsBodyT

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			endSession()
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			endSession()
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		createNewChat := func() error {

			helpers.MapToStruct(body.Data, &newChatBody)

			if val_err := newChatBody.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			data, app_err := groupChatService.NewGroupChat(
				newChatBody.Name,
				newChatBody.Description,
				newChatBody.PictureData,
				[]string{fmt.Sprint(clientUser.Id), clientUser.Username},
				newChatBody.InitUsers,
			)
			if app_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			}

			return c.WriteJSON(data)
		}

		// For Group chat, messages can only be acknowledged in batches,
		acknowledgeMessages := func() error {

			helpers.MapToStruct(body.Data, &ackMsgsBody)

			if val_err := ackMsgsBody.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			go groupChatService.BatchUpdateMessageDeliveryStatus(ackMsgsBody.GroupChatId, clientUser.Id, ackMsgsBody.Status, ackMsgsBody.MsgAckDatas)

			return nil
		}

		if body.Action == "create new chat" {

			w_err = createNewChat()

		} else if body.Action == "acknowledge messages" {

			w_err = acknowledgeMessages()

		} else {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("invalid 'action' value")))
		}
	}
}

// This handler merely get chats as is from the database, no updates accounted for yet.
// After closing this,  we must immediately access "Open[DM|Group]ChatStream"
var GetMyChats = helpers.WSHandlerProtected(func(c *websocket.Conn) {

	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	myChats, app_err := user.GetChats(clientUser.Id)

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusInternalServerError, app_err))
	} else {
		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       myChats,
		})
	}

	if w_err != nil {
		log.Println(w_err)
		return
	}
})

var GetAllUsers = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	allUsers, app_err := user.GetAll(clientUser.Id)

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
	} else {
		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       allUsers,
		})
	}

	if w_err != nil {
		log.Println(w_err)
		return
	}
})

var SearchUser = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var body struct {
		Query string
	}

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			return
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			return
		}

		searchResult, app_err := user.Search(clientUser.Id, body.Query)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       searchResult,
		})
	}
})

var FindNearbyUsers = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var body findNearbyUsersBody

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			return
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			return
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		nearbyUsers, app_err := user.FindNearby(clientUser.Id, body.LiveLocation)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       nearbyUsers,
		})
	}
})

var SwitchMyPresence = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var body switchMyPresenceBody

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			return
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			return
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		app_err := userService.SwitchMyPresence(clientUser.Id, body.Presence, body.LastSeen)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body: map[string]any{
				"msg": "Operation Successful",
			},
		})

	}
})

var UpdateMyGeolocation = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var body updateMyGeolocationBody

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			return
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			return
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		app_err := user.UpdateLocation(clientUser.Id, body.NewGeolocation)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body: map[string]any{
				"msg": "Operation Successful",
			},
		})
	}
})
