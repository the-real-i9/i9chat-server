package userControllers

import (
	"fmt"
	"i9chat/services/appObservers"
	"i9chat/services/appServices"
	"i9chat/services/chatService"
	"i9chat/services/userService"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var ChangeProfilePicture = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		PictureData []byte
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			// log.Println(r_err)
			break
		}

		var w_err error
		if app_err := (userService.User{Id: user.UserId}).ChangeProfilePicture(body.PictureData); app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("Operation failed: %s", app_err)))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": fiber.StatusOK,
				"body": map[string]any{
					"msg": "Operation Successful",
				},
			})
		}

		if w_err != nil {
			// log.Println(w_err)
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
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	// a channel for streaming data to client
	var myMailbox = make(chan map[string]any, 5)

	// subscribe to receiving chat updates
	// myMailbox is passed by reference to an observer keeping several mailboxes wanting to receive updates
	dco := appObservers.DMChatObserver{}

	mailboxKey := fmt.Sprintf("user-%d", user.UserId)

	dco.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		dco.Unsubscribe(mailboxKey)
	}

	go createNewDMChatAndAckMessages(c, user, endSession)

	/* ---- stream chat events pending dispatch to the myMailbox ---- */
	// this excecutes once every connection restoration
	// "What did I miss?" - When the client comes online they get a report of all missed data (while offline)
	if eventDataList, err := (userService.User{Id: user.UserId}).GetDMChatEventsPendingReceipt(); err == nil {
		for _, eventData := range eventDataList {
			eventData := *eventData
			myMailbox <- eventData
		}
	}

	// when myMailbox receives any new data, it is sent to the client
	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			// log.Println(w_err)
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
func createNewDMChatAndAckMessages(c *websocket.Conn, user appTypes.JWTUserData, endSession func()) {
	var body struct {
		Action string
		Data   map[string]any
	}

	var newChatBody struct {
		PartnerId int
		Msg       map[string]any
		CreatedAt time.Time
	}

	// For DM Chat, we allowed options for both single and batch acknowledgements
	var ackMsgBody struct {
		Status string
		appTypes.DMChatMsgDeliveryData
	}

	var batchAckMsgBody struct {
		Status   string
		MsgDatas []*appTypes.DMChatMsgDeliveryData
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			// log.Println(r_err)
			endSession()
			return
		}

		createNewChat := func() error {

			helpers.MapToStruct(body.Data, &newChatBody)

			var w_err error

			data, app_err := chatService.NewDMChat(
				user.UserId,
				newChatBody.PartnerId,
				appServices.MessageBinaryToUrl(user.UserId, newChatBody.Msg),
				newChatBody.CreatedAt,
			)
			if app_err != nil {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			} else {
				w_err = c.WriteJSON(data)
			}

			if w_err != nil {
				return w_err
			}

			return nil
		}

		// acknowledge messages singly
		acknowledgeMessage := func() {

			helpers.MapToStruct(body.Data, &ackMsgBody)

			go chatService.DMChatMessage{
				Id:       ackMsgBody.MsgId,
				DmChatId: ackMsgBody.DmChatId,
				SenderId: ackMsgBody.SenderId,
			}.UpdateDeliveryStatus(user.UserId, ackMsgBody.Status, ackMsgBody.At)
		}

		// acknowledge messages in batch
		batchAcknowledgeMessages := func() {

			helpers.MapToStruct(body.Data, &batchAckMsgBody)

			go chatService.BatchUpdateDMChatMessageDeliveryStatus(user.UserId, batchAckMsgBody.Status, batchAckMsgBody.MsgDatas)
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
				// log.Println(w_err)
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
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	// a channel for streaming data to client
	var myMailbox = make(chan map[string]any, 5)

	// subscribe to receiving chat updates
	// myMailbox is passed by reference to an observer keeping several mailboxes wanting to receive updates
	gco := appObservers.GroupChatObserver{}

	mailboxKey := fmt.Sprintf("user-%d", user.UserId)

	gco.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		gco.Unsubscribe(mailboxKey)
	}

	go createNewGroupChatAndAckMessages(c, user, endSession)

	/* ---- stream chat events pending dispatch to the myMailbox ---- */
	// this excecutes once every connection restoration
	// "What did I miss?" - When the client comes online they get a report of all missed data (while offline)
	if eventDataList, err := (userService.User{Id: user.UserId}).GetGroupChatEventsPendingReceipt(); err == nil {
		for _, eventData := range eventDataList {
			eventData := *eventData
			myMailbox <- eventData
		}
	}

	// when myMailbox receives any new data, it is sent to the client
	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			// log.Println(w_err)
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
func createNewGroupChatAndAckMessages(c *websocket.Conn, user appTypes.JWTUserData, endSession func()) {
	var body struct {
		Action string
		Data   map[string]any
	}

	var newChatBody struct {
		Name        string
		Description string
		PictureData []byte
		InitUsers   [][]string
	}

	// For Group chat, messages can only be acknowledged in batches,
	// and it's one group chat and one status at a time
	var ackMsgsBody struct {
		Status      string
		GroupChatId int
		MsgDatas    []*appTypes.GroupChatMsgDeliveryData
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			// log.Println(r_err)
			endSession()
			return
		}

		createNewChat := func() error {

			helpers.MapToStruct(body.Data, &newChatBody)

			var w_err error

			data, app_err := chatService.NewGroupChat(
				newChatBody.Name,
				newChatBody.Description,
				newChatBody.PictureData,
				[]string{fmt.Sprint(user.UserId), user.Username},
				newChatBody.InitUsers,
			)
			if app_err != nil {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			} else {
				w_err = c.WriteJSON(data)
			}

			if w_err != nil {
				// log.Println(w_err)
			}

			return nil
		}

		// For Group chat, messages can only be acknowledged in batches,
		acknowledgeMessages := func() {

			helpers.MapToStruct(body.Data, &ackMsgsBody)

			go chatService.GroupChat{Id: ackMsgsBody.GroupChatId}.BatchUpdateGroupChatMessageDeliveryStatus(user.UserId, ackMsgsBody.Status, ackMsgsBody.MsgDatas)
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
				// log.Println(w_err)
				endSession()
				return
			}
		}
	}
}

var GetMyChats = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	// This handler merely get chats as is from the database, no updates accounted for yet
	// After this guy closes:
	// We must "Init[DM|Group]ChatStream"

	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	myChats, app_err := userService.User{Id: user.UserId}.GetMyChats()

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.ErrResp(500, app_err))
	} else {
		w_err = c.WriteJSON(map[string]any{
			"statusCode": fiber.StatusOK,
			"body":       myChats,
		})
	}

	if w_err != nil {
		// log.Println(w_err)
		return
	}
})

var GetAllUsers = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	allUsers, app_err := userService.GetAllUsers(user.UserId)

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
		// log.Println(w_err)
		return
	}
})

var SearchUser = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		Query string
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			return
		}

		searchResult, app_err := userService.SearchUser(user.UserId, body.Query)

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
			// log.Println(w_err)
			return
		}
	}
})

var FindNearbyUsers = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		LiveLocation string
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			return
		}

		nearbyUsers, app_err := userService.FindNearbyUsers(user.UserId, body.LiveLocation)

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
			// log.Println(w_err)
			return
		}
	}
})
