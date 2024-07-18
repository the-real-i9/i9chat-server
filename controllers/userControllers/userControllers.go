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
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

var ChangeProfilePicture = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var body struct {
		PictureData []byte
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		var w_err error
		if app_err := userService.ChangeMyProfilePicture(clientUser.Id, body.PictureData); app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("operation failed: %s", app_err)))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": fiber.StatusOK,
				"body": map[string]any{
					"msg": "Operation Successful",
				},
			})
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}
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
			return
		}
	}
})

// This goroutine handles:
//
// + initating new dm chats
//
// + sending acknowledgements for received dm messages
func createNewDMChatAndAckMessages(c *websocket.Conn, clientUser *appTypes.ClientUser, endSession func()) {
	var body struct {
		Action string
		Data   map[string]any
	}

	var newChatBody struct {
		PartnerId int
		InitMsg   map[string]any
		CreatedAt time.Time
	}

	// For DM Chat, we allowed options for both single and batch acknowledgements
	var ackMsgBody struct {
		Status string
		*appTypes.DMChatMsgAckData
	}

	var batchAckMsgBody struct {
		Status      string
		MsgAckDatas []*appTypes.DMChatMsgAckData
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			endSession()
			return
		}

		createNewChat := func() error {

			helpers.MapToStruct(body.Data, &newChatBody)

			var w_err error

			initiatorData, app_err := dmChatService.NewDMChat(
				clientUser.Id,
				newChatBody.PartnerId,
				appServices.MessageBinaryToUrl(clientUser.Id, newChatBody.InitMsg),
				newChatBody.CreatedAt,
			)
			if app_err != nil {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			} else {
				w_err = c.WriteJSON(initiatorData)
			}

			if w_err != nil {
				return w_err
			}

			return nil
		}

		// acknowledge messages singly
		acknowledgeMessage := func() {

			helpers.MapToStruct(body.Data, &ackMsgBody)

			go dmChatService.UpdateMessageDeliveryStatus(ackMsgBody.DMChatId, ackMsgBody.MsgId, ackMsgBody.SenderId, clientUser.Id, ackMsgBody.Status, ackMsgBody.At)
		}

		// acknowledge messages in batch
		batchAcknowledgeMessages := func() {

			helpers.MapToStruct(body.Data, &batchAckMsgBody)

			go dmChatService.BatchUpdateMessageDeliveryStatus(clientUser.Id, batchAckMsgBody.Status, batchAckMsgBody.MsgAckDatas)
		}

		if body.Action == "create new chat" {

			if err := createNewChat(); err != nil {
				log.Println(err)
				endSession()
				return
			}
		} else if body.Action == "acknowledge message" {

			acknowledgeMessage()

		} else if body.Action == "batch acknowledge messages" {

			batchAcknowledgeMessages()

		} else {
			if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("invalid 'action' value"))); w_err != nil {
				log.Println(w_err)
				endSession()
				return
			}
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
	var body struct {
		Action string
		Data   map[string]any
	}

	var newChatBody struct {
		Name        string
		Description string
		PictureData []byte
		InitUsers   [][]appTypes.String
	}

	// For Group chat, messages should be acknowledged in batches,
	// and it's only for a single group chat at a time
	var ackMsgsBody struct {
		Status      string
		GroupChatId int
		MsgAckDatas []*appTypes.GroupChatMsgAckData
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			endSession()
			return
		}

		createNewChat := func() error {

			helpers.MapToStruct(body.Data, &newChatBody)

			var w_err error

			data, app_err := groupChatService.NewGroupChat(
				newChatBody.Name,
				newChatBody.Description,
				newChatBody.PictureData,
				[]string{fmt.Sprint(clientUser.Id), clientUser.Username},
				newChatBody.InitUsers,
			)
			if app_err != nil {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			} else {
				w_err = c.WriteJSON(data)
			}

			if w_err != nil {
				log.Println(w_err)
			}

			return nil
		}

		// For Group chat, messages can only be acknowledged in batches,
		acknowledgeMessages := func() {

			helpers.MapToStruct(body.Data, &ackMsgsBody)

			go groupChatService.BatchUpdateMessageDeliveryStatus(ackMsgsBody.GroupChatId, clientUser.Id, ackMsgsBody.Status, ackMsgsBody.MsgAckDatas)
		}

		if body.Action == "create new chat" {

			if err := createNewChat(); err != nil {
				log.Println(err)
				endSession()
				return
			}
		} else if body.Action == "acknowledge messages" {

			acknowledgeMessages()
		} else {
			if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("invalid 'action' value"))); w_err != nil {
				log.Println(w_err)
				endSession()
				return
			}
		}
	}
}

// This handler merely get chats as is from the database, no updates accounted for yet.
// After closing this,  we must immediately access "Init[DM|Group]ChatStream"
var GetMyChats = helpers.WSHandlerProtected(func(c *websocket.Conn) {

	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	myChats, app_err := user.GetChats(clientUser.Id)

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusInternalServerError, app_err))
	} else {
		w_err = c.WriteJSON(map[string]any{
			"statusCode": fiber.StatusOK,
			"body":       myChats,
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
		w_err = c.WriteJSON(map[string]any{
			"statusCode": fiber.StatusOK,
			"body":       allUsers,
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

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			return
		}

		searchResult, app_err := user.Search(clientUser.Id, body.Query)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": 200,
				"body":       searchResult,
			})
		}

		if w_err != nil {
			log.Println(w_err)
			return
		}
	}
})

var FindNearbyUsers = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var body struct {
		LiveLocation string
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			return
		}

		nearbyUsers, app_err := user.FindNearby(clientUser.Id, body.LiveLocation)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": 200,
				"body":       nearbyUsers,
			})
		}

		if w_err != nil {
			log.Println(w_err)
			return
		}
	}
})

var SwitchMyPresence = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var body struct {
		Presence string
		LastSeen pgtype.Timestamp
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			return
		}

		app_err := userService.SwitchMyPresence(clientUser.Id, body.Presence, body.LastSeen)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": 200,
				"body": map[string]any{
					"msg": "Operation Successful",
				},
			})
		}

		if w_err != nil {
			log.Println(w_err)
			return
		}
	}
})

var UpdateMyGeolocation = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var body struct {
		NewGeolocation string
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			return
		}

		app_err := user.UpdateLocation(clientUser.Id, body.NewGeolocation)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": 200,
				"body": map[string]any{
					"msg": "Operation Successful",
				},
			})
		}

		if w_err != nil {
			log.Println(w_err)
			return
		}
	}
})
