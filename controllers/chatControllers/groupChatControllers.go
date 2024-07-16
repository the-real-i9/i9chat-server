package chatControllers

import (
	"fmt"
	user "i9chat/models/userModel"
	"i9chat/services/appObservers"
	"i9chat/services/appServices"
	"i9chat/services/chatService"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetGroupChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var clientUser appTypes.ClientUser

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &clientUser)

	var body struct {
		GroupChatId int
		Offset      int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		log.Println(r_err)
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
		log.Println(w_err)
		return
	}
})

// this goroutine receives message acknowlegement for sent messages,
// and in turn changes the delivery status of messages sent by the child goroutine
var OpenGroupMessagingStream = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var clientUser appTypes.ClientUser

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &clientUser)

	var groupChatId int

	fmt.Sscanf(c.Params("group_chat_id"), "%d", &groupChatId)

	var myMailbox = make(chan map[string]any, 2)

	mailboxKey := fmt.Sprintf("user-%d--groupchat-%d", clientUser.Id, groupChatId)

	gcso := appObservers.GroupChatSessionObserver{}

	gcso.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		gcso.Unsubscribe(mailboxKey)
	}

	go sendGroupChatMessages(c, clientUser, groupChatId, endSession)

	/* ---- stream group chat message events pending dispatch to the channel ---- */
	// observe that this happens once every new connection
	// A "What did I miss?" sort of query
	if eventDataList, err := user.GetGroupChatMessageEventsPendingReceipt(clientUser.Id, groupChatId); err == nil {
		for _, eventData := range eventDataList {
			eventData := *eventData
			myMailbox <- eventData
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

func sendGroupChatMessages(c *websocket.Conn, clientUser appTypes.ClientUser, groupChatId int, endSession func()) {
	// this goroutine sends messages
	var body struct {
		Msg map[string]any
		At  time.Time
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		data, app_err := chatService.GroupChat{Id: groupChatId}.SendMessage(
			clientUser.Id,
			appServices.MessageBinaryToUrl(clientUser.Id, body.Msg),
			body.At,
		)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(data)
		}

		if w_err != nil {
			log.Println(w_err)
			endSession()
			return
		}
	}
}

var ExecuteGroupAction = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var clientUser appTypes.ClientUser

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &clientUser)

	type action string
	type handler func(clientUser []string, data map[string]any) error

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
			log.Println(r_err)
			break
		}

		app_err := operationToHandlerMap[body.Action]([]string{fmt.Sprint(clientUser.Id), clientUser.Username}, body.Data)

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
			log.Println(w_err)
			break
		}
	}
})

func changeGroupName(clientUser []string, data map[string]any) error {
	var d struct {
		GroupChatId int
		NewName     string
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.ChangeName(clientUser, d.NewName)

}

func changeGroupDescription(clientUser []string, data map[string]any) error {
	var d struct {
		GroupChatId    int
		NewDescription string
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.ChangeDescription(clientUser, d.NewDescription)
}

func changeGroupPicture(clientUser []string, data map[string]any) error {
	var d struct {
		GroupChatId    int
		NewPictureData []byte
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.ChangePicture(clientUser, d.NewPictureData)
}

func addUsersToGroup(clientUser []string, data map[string]any) error {
	var d struct {
		GroupChatId int
		NewUsers    [][]appTypes.String
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.AddUsers(clientUser, d.NewUsers)
}

func removeUserFromGroup(clientUser []string, data map[string]any) error {
	var d struct {
		GroupChatId int
		User        []appTypes.String
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.RemoveUser(clientUser, d.User)
}

func joinGroup(clientUser []string, data map[string]any) error {
	var d struct {
		GroupChatId int
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.Join(clientUser)
}

func leaveGroup(clientUser []string, data map[string]any) error {
	var d struct {
		GroupChatId int
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.Leave(clientUser)
}

func makeUserGroupAdmin(clientUser []string, data map[string]any) error {
	var d struct {
		GroupChatId int
		User        []appTypes.String
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.MakeUserAdmin(clientUser, d.User)
}

func removeUserFromGroupAdmins(clientUser []string, data map[string]any) error {
	var d struct {
		GroupChatId int
		User        []appTypes.String
	}

	helpers.MapToStruct(data, &d)

	return chatService.GroupChat{Id: d.GroupChatId}.RemoveUserFromAdmins(clientUser, d.User)
}
