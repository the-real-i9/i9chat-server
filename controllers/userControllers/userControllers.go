package usercontrollers

import (
	"fmt"
	"log"
	"services/appservices"
	"services/chatservice"
	"services/userservice"
	"time"
	"utils/appglobals"
	"utils/apptypes"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var ChangeProfilePicture = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.ParseToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		Picture []byte
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		var w_err error
		if app_err := (userservice.User{Id: user.UserId}).ChangeProfilePicture(body.Picture); app_err != nil {
			w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(map[string]any{"code": fiber.StatusOK, "msg": "Operation Successful"})
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}
	}
})

var InitDMChatStream = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	// this goroutine streams chat updates to the client
	// including new chats and new messages

	var user apptypes.JWTUserData

	helpers.ParseToStruct(c.Locals("auth").(map[string]any), &user)

	// a data channel for transmitting data
	var myMailbox = make(chan map[string]any, 5)

	// subscribe to receiving chat updates
	// myMailbox is passed by reference to an observer keeping several mailboxes wanting to receive updates
	nco := appglobals.DMChatObserver{}

	mailboxKey := fmt.Sprintf("user-%d", user.UserId)

	nco.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		nco.Unsubscribe(mailboxKey)
	}

	go createNewDMChatAndAckMessages(c, user, endSession)

	/* ---- stream chat events pending dispatch to the channel ---- */
	// observe that this happens once every new connection
	// A "What did I miss?" sort of query
	if event_data_kvps, err := (userservice.User{Id: user.UserId}).GetDMChatEventsPendingDispatch(); err == nil {
		for _, evk := range event_data_kvps {
			evk := *evk
			myMailbox <- evk
		}
	}

	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			log.Println(w_err)
			endSession()
			return
		}
	}
})

func createNewDMChatAndAckMessages(c *websocket.Conn, user apptypes.JWTUserData, endSession func()) {
	var body struct {
		Action string
		Data   map[string]any
	}

	var newChatBody struct {
		PartnerId int
		Msg       map[string]any
		CreatedAt time.Time
	}

	// For DM Chat, you have the options for both single and batch acknowledgements
	var ackMsgBody struct {
		Status   string
		MsgId    int
		ChatId   int
		SenderId int
		At       time.Time
	}

	var batchAckMsgBody struct {
		Status   string
		MsgDatas []*apptypes.DMChatMsgDeliveryData
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			endSession()
			return
		}

		createNewChat := func() error {
			helpers.ParseToStruct(body.Data, &newChatBody)

			var w_err error
			data, app_err := chatservice.NewDMChat(
				user.UserId,
				newChatBody.PartnerId,
				appservices.MessageBinaryToUrl(user.UserId, newChatBody.Msg),
				newChatBody.CreatedAt,
			)
			if app_err != nil {
				w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
			} else {
				w_err = c.WriteJSON(data)
			}

			if w_err != nil {
				return w_err
			}

			return nil
		}

		acknowledgeMessage := func() {
			helpers.ParseToStruct(body.Data, &ackMsgBody)

			go chatservice.DMChatMessage{
				Id:       ackMsgBody.MsgId,
				DmChatId: ackMsgBody.ChatId,
				SenderId: ackMsgBody.SenderId,
			}.UpdateDeliveryStatus(user.UserId, ackMsgBody.Status, ackMsgBody.At)
		}

		batchAcknowledgeMessages := func() {
			helpers.ParseToStruct(body.Data, &batchAckMsgBody)

			go chatservice.BatchUpdateDMChatMessageDeliveryStatus(user.UserId, batchAckMsgBody.Status, batchAckMsgBody.MsgDatas)
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
			if w_err := c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, fmt.Errorf("invalid 'action' value"))); w_err != nil {
				log.Println(w_err)
				endSession()
				return
			}
		}
	}
}

var InitGroupChatStream = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	// this goroutine streams chat updates to the client
	// including new chats and new messages
	var user apptypes.JWTUserData

	helpers.ParseToStruct(c.Locals("auth").(map[string]any), &user)

	// a data channel for transmitting data
	var myMailbox = make(chan map[string]any, 5)

	// subscribe to receiving chat updates
	// myMailbox is passed by reference to an observer keeping several mailboxes wanting to receive updates
	nco := appglobals.GroupChatObserver{}

	mailboxKey := fmt.Sprintf("user-%d", user.UserId)

	nco.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		nco.Unsubscribe(mailboxKey)
	}

	go createNewGroupDMChatAndAckMessages(c, user, endSession)

	/* ---- stream chat events pending dispatch to the channel ---- */
	// observe that this happens once every new connection
	// A "What did I miss?" sort of query
	if event_data_kvps, err := (userservice.User{Id: user.UserId}).GetGroupChatEventsPendingDispatch(); err == nil {
		for _, evk := range event_data_kvps {
			evk := *evk
			myMailbox <- evk
		}
	}

	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			log.Println(w_err)
			endSession()
			return
		}
	}
})

func createNewGroupDMChatAndAckMessages(c *websocket.Conn, user apptypes.JWTUserData, endSession func()) {
	var body struct {
		Action string
		Data   map[string]any
	}

	var newChatBody struct {
		Name        string
		Description string
		Picture     []byte
		InitUsers   [][]string
	}

	// For Group chat, messages can only be acknowledged in batches,
	// and it's one group chat and one status at a time
	var ackMsgsBody struct {
		ChatId   int
		Status   string
		MsgDatas []*apptypes.GroupChatMsgDeliveryData
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			endSession()
			return
		}

		createNewChat := func() error {
			helpers.ParseToStruct(body.Data, &newChatBody)

			var w_err error
			data, app_err := chatservice.NewGroupChat(
				newChatBody.Name, newChatBody.Description, newChatBody.Picture,
				[]string{fmt.Sprint(user.UserId), user.Username}, newChatBody.InitUsers,
			)
			if app_err != nil {
				w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
			} else {
				w_err = c.WriteJSON(data)
			}

			if w_err != nil {
				log.Println(w_err)
			}

			return nil
		}

		acknowledgeMessages := func() {
			helpers.ParseToStruct(body.Data, &ackMsgsBody)

			go chatservice.GroupChat{Id: ackMsgsBody.ChatId}.BatchUpdateGroupChatMessageDeliveryStatus(user.UserId, ackMsgsBody.Status, ackMsgsBody.MsgDatas)
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
			if w_err := c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, fmt.Errorf("invalid 'action' value"))); w_err != nil {
				log.Println(w_err)
				endSession()
				return
			}
		}
	}
}

var GetMyChats = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	// This guy merely get chats as is in the database, no updates accounted for yet
	// After this guy closes:
	// We must "InitChatUpdateStream"

	var user apptypes.JWTUserData

	helpers.ParseToStruct(c.Locals("auth").(map[string]any), &user)

	myChats, app_err := userservice.User{Id: user.UserId}.GetMyChats()

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

var GetAllUsers = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.ParseToStruct(c.Locals("auth").(map[string]any), &user)

	allUsers, app_err := userservice.User{Id: user.UserId}.GetAllUsers()

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(map[string]any{
			"statusCode": 200,
			"body": map[string]any{
				"all_users": make([]any, 0),
			},
		})
	} else {
		w_err = c.WriteJSON(map[string]any{
			"statusCode": 200,
			"body": map[string]any{
				"all_users": allUsers,
			},
		})
	}

	if w_err != nil {
		log.Println(w_err)
		return
	}
})