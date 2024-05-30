package chatControllers

import (
	"fmt"
	"i9chat/services/appObservers"
	"i9chat/services/appServices"
	"i9chat/services/chatService"
	"i9chat/services/userService"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetGroupChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		GroupChatId int
		Offset      int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		// log.Println(r_err)
		return
	}

	groupChatHistory, app_err := chatService.GroupChat{Id: body.GroupChatId}.GetChatHistory(body.Offset)

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
	} else {
		w_err = c.WriteJSON(map[string]any{
			"statusCode": 200,
			"body":       groupChatHistory,
		})
	}

	if w_err != nil {
		// log.Println(w_err)
		return
	}
})

// this goroutine receives message acknowlegement for sent messages,
// and in turn changes the delivery status of messages sent by the child goroutine
var OpenGroupMessagingStream = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var groupChatId int

	fmt.Sscanf(c.Params("group_chat_id"), "%d", &groupChatId)

	var myMailbox = make(chan map[string]any, 2)

	mailboxKey := fmt.Sprintf("user-%d--groupchat-%d", user.UserId, groupChatId)

	gcso := appObservers.GroupChatSessionObserver{}

	gcso.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		gcso.Unsubscribe(mailboxKey)
	}

	go sendGroupChatMessages(c, user, groupChatId, endSession)

	/* ---- stream group chat message events pending dispatch to the channel ---- */
	// observe that this happens once every new connection
	// A "What did I miss?" sort of query
	if eventDataList, err := (userService.User{Id: user.UserId}).GetGroupChatMessageEventsPendingReceipt(groupChatId); err == nil {
		for _, eventData := range eventDataList {
			eventData := *eventData
			myMailbox <- eventData
		}
	}

	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			// log.Println(w_err)
			endSession()
			return
		}
	}
})

func sendGroupChatMessages(c *websocket.Conn, user appTypes.JWTUserData, groupChatId int, endSession func()) {
	// this goroutine sends messages
	var body struct {
		Msg map[string]any
		At  time.Time
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			// log.Println(r_err)
			return
		}

		data, app_err := chatService.GroupChat{Id: groupChatId}.SendMessage(
			user.UserId,
			appServices.MessageBinaryToUrl(user.UserId, body.Msg),
			body.At,
		)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(data)
		}

		if w_err != nil {
			// log.Println(w_err)
			endSession()
			return
		}
	}
}

var ExecuteGroupAction = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user appTypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	type action string
	type handler func(client []string, data map[string]any) error

	operationToHandlerMap := map[action]handler{
		"change_name":             changeGroupName,
		"change_description":      changeGroupDescription,
		"change_picture":          changeGroupPicture,
		"add_users":               addUsersToGroup,
		"remove_user":             removeUserFromGroup,
		"join":                    joinGroup,
		"leave":                   leaveGroup,
		"make_user_admin":         makeUserGroupAdmin,
		"remove_user_from_admins": removeUserFromGroupAdmins,
	}

	for {
		var body struct {
			Action action
			Data   map[string]any
		}

		if r_err := c.ReadJSON(&body); r_err != nil {
			// log.Println(r_err)
			break
		}

		client := []string{fmt.Sprint(user.UserId), user.Username}

		app_err := operationToHandlerMap[body.Action](client, body.Data)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("operation failed: %s", app_err)))
		} else {
			w_err = c.WriteJSON(map[string]any{
				"statusCode": 200,
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

func changeGroupName(client []string, data map[string]any) error {
	var d struct {
		GroupChatId int
		NewName     string
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.ChangeName(client, d.NewName)

}

func changeGroupDescription(client []string, data map[string]any) error {
	var d struct {
		GroupChatId    int
		NewDescription string
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.ChangeDescription(client, d.NewDescription)
}

func changeGroupPicture(client []string, data map[string]any) error {
	var d struct {
		GroupChatId    int
		NewPictureData []byte
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.ChangePicture(client, d.NewPictureData)
}

func addUsersToGroup(client []string, data map[string]any) error {
	var d struct {
		GroupChatId int
		NewUsers    [][]appTypes.String
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.AddUsers(client, d.NewUsers)
}

func removeUserFromGroup(client []string, data map[string]any) error {
	var d struct {
		GroupChatId int
		User        []appTypes.String
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.RemoveUser(client, d.User)
}

func joinGroup(client []string, data map[string]any) error {
	var d struct {
		GroupChatId int
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.Join(client)
}

func leaveGroup(client []string, data map[string]any) error {
	var d struct {
		GroupChatId int
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.Leave(client)
}

func makeUserGroupAdmin(client []string, data map[string]any) error {
	var d struct {
		GroupChatId int
		User        []appTypes.String
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.MakeUserAdmin(client, d.User)
}

func removeUserFromGroupAdmins(client []string, data map[string]any) error {
	var d struct {
		GroupChatId int
		User        []appTypes.String
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.RemoveUserFromAdmins(client, d.User)
}
