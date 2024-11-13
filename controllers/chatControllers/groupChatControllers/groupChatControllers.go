package groupChatControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/services/chatServices/groupChatService"
	"i9chat/services/utils/authUtilServices"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var CreateNewGroupChatAndAckMessages = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var w_err error

	for {
		var body createNewGroupChatAndAckMessagesBody

		var newChatData newGroupChatDataT

		// For Group chat, messages should be acknowledged in batches,
		// and it's only for a single group chat at a time
		var ackMsgsData ackMsgsDataT

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

		createNewChat := func() error {

			helpers.MapToStruct(body.Data, &newChatData)

			if val_err := newChatData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			respData, app_err := groupChatService.NewGroupChat(
				newChatData.Name,
				newChatData.Description,
				newChatData.PictureData,
				[]string{fmt.Sprint(clientUser.Id), clientUser.Username},
				newChatData.InitUsers,
			)
			if app_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			}

			return c.WriteJSON(respData)
		}

		// For Group chat, messages can only be acknowledged in batches,
		acknowledgeMessages := func() error {

			helpers.MapToStruct(body.Data, &ackMsgsData)

			if val_err := ackMsgsData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			go groupChatService.BatchUpdateMessageDeliveryStatus(ackMsgsData.GroupChatId, clientUser.Id, ackMsgsData.Status, ackMsgsData.MsgAckDatas)

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
})

var GetChatHistory = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {

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

		respData, app_err := groupChatService.GetChatHistory(body.GroupChatId, body.Offset)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       respData,
		})
	}
})

var SendMessage = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var groupChatId int

	_, err := fmt.Sscanf(c.Params("group_chat_id"), "%d", &groupChatId)
	if err != nil {
		panic(err)
	}

	var w_err error

	for {
		var body sendMessageBody

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

		respData, app_err := groupChatService.SendMessage(
			groupChatId,
			clientUser.Id,
			body.Msg,
			body.At,
		)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(respData)

	}
})

var ExecuteAction = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

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
				"msg": "operation successful",
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
