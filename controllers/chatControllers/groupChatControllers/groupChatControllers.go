package groupChatControllers

import (
	"fmt"
	groupChat "i9chat/models/chatModel/groupChatModel"
	user "i9chat/models/userModel"
	"i9chat/services/appObservers"
	"i9chat/services/appServices"
	"i9chat/services/chatService/groupChatService"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {

	var w_err error

	for {
		var body getChatHistoryBody

		if w_err != nil {
			log.Println(w_err)
			return
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		groupChatHistory, app_err := groupChat.GetChatHistory(body.GroupChatId, body.Offset)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       groupChatHistory,
		})
	}
})

// this goroutine receives message acknowlegement for sent messages,
// and in turn changes the delivery status of messages sent by the child goroutine
var OpenMessagingStream = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	var groupChatId int

	_, param_err := fmt.Sscanf(c.Params("group_chat_id"), "%d", &groupChatId)
	if param_err != nil {
		if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusBadRequest, fmt.Errorf("parameter group_chat_id is not a number"))); w_err != nil {
			log.Println(w_err)
		}
		return
	}

	var myMailbox = make(chan map[string]any, 2)

	mailboxKey := fmt.Sprintf("user-%d--groupchat-%d", clientUser.Id, groupChatId)

	gcso := appObservers.GroupChatSessionObserver{}

	gcso.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		gcso.Unsubscribe(mailboxKey)
	}

	go sendMessages(c, clientUser, groupChatId, endSession)

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

func sendMessages(c *websocket.Conn, clientUser *appTypes.ClientUser, groupChatId int, endSession func()) {
	// this goroutine sends messages

	var w_err error

	for {
		var body openMessagingStreamBody

		if w_err != nil {
			log.Println(w_err)
			endSession()
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

		senderData, app_err := groupChatService.SendMessage(
			groupChatId,
			clientUser.Id,
			appServices.MessageBinaryToUrl(clientUser.Id, body.Msg),
			body.At,
		)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(senderData)

	}
}

var ExecuteAction = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("auth").(*appTypes.ClientUser)

	type handler func(clientUser []string, data map[string]any) error

	actionToHandlerMap := map[action]handler{
		"change name":             changeGroupName,
		"change description":      changeGroupDescription,
		"change picture":          changeGroupPicture,
		"add users":               addUsersToGroup,
		"remove user":             removeUserFromGroup,
		"join":                    joinGroup,
		"leave":                   leaveGroup,
		"make user admin":         makeUserGroupAdmin,
		"remove user from admins": removeUserFromGroupAdmins,
	}

	var w_err error

	for {
		var body executeActionBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		if r_err := c.ReadJSON(&body); r_err != nil {
			log.Println(r_err)
			break
		}

		app_err := actionToHandlerMap[body.Action]([]string{fmt.Sprint(clientUser.Id), clientUser.Username}, body.Data)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("action failed: %s", app_err)))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body: map[string]any{
				"msg": "Action executed",
			},
		})
	}
})

func changeGroupName(clientUser []string, data map[string]any) error {
	var d changeGroupNameT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.ChangeGroupName(d.GroupChatId, clientUser, d.NewName)

}

func changeGroupDescription(clientUser []string, data map[string]any) error {
	var d changeGroupDescriptionT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.ChangeGroupDescription(d.GroupChatId, clientUser, d.NewDescription)
}

func changeGroupPicture(clientUser []string, data map[string]any) error {
	var d changeGroupPictureT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.ChangeGroupPicture(d.GroupChatId, clientUser, d.NewPictureData)
}

func addUsersToGroup(clientUser []string, data map[string]any) error {
	var d addUsersToGroupT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.AddUsersToGroup(d.GroupChatId, clientUser, d.NewUsers)
}

func removeUserFromGroup(clientUser []string, data map[string]any) error {
	var d actOnSingleUserT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.RemoveUserFromGroup(d.GroupChatId, clientUser, d.User)
}

func joinGroup(clientUser []string, data map[string]any) error {
	var d joinLeaveGroupT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.JoinGroup(d.GroupChatId, clientUser)
}

func leaveGroup(clientUser []string, data map[string]any) error {
	var d joinLeaveGroupT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.LeaveGroup(d.GroupChatId, clientUser)
}

func makeUserGroupAdmin(clientUser []string, data map[string]any) error {
	var d actOnSingleUserT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.MakeUserGroupAdmin(d.GroupChatId, clientUser, d.User)
}

func removeUserFromGroupAdmins(clientUser []string, data map[string]any) error {
	var d actOnSingleUserT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.RemoveUserFromGroupAdmins(d.GroupChatId, clientUser, d.User)
}
