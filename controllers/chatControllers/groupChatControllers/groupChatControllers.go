package groupChatControllers

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/services/chatServices/groupChatService"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func CreateNewGroupChat(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var body newGroupChatBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, app_err := groupChatService.NewGroupChat(ctx,
		body.Name,
		body.Description,
		body.PictureData,
		[]string{fmt.Sprint(clientUser.Id), clientUser.Username},
		body.InitUsers,
	)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

var GetChatHistory = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
			w_err = c.WriteJSON(helpers.ErrResp(val_err))
			continue
		}

		respData, app_err := groupChatService.GetChatHistory(ctx, body.GroupChatId, body.Offset)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       respData,
		})
	}
})

var SendMessage = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
			w_err = c.WriteJSON(helpers.ErrResp(val_err))
			continue
		}

		respData, app_err := groupChatService.SendMessage(ctx,
			groupChatId,
			clientUser.Id,
			body.Msg,
			body.At,
		)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(app_err))
			continue
		}

		w_err = c.WriteJSON(respData)

	}
})

var ExecuteAction = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	type handler func(ctx context.Context, clientUser []string, data map[string]any) error

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

		app_err := actionToHandlerMap[body.Action](ctx, []string{fmt.Sprint(clientUser.Id), clientUser.Username}, body.Data)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fmt.Errorf("action failed: %s", app_err)))
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

func changeGroupName(ctx context.Context, clientUser []string, data map[string]any) error {
	var d changeGroupNameT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.ChangeGroupName(ctx, d.GroupChatId, clientUser, d.NewName)
}

func changeGroupDescription(ctx context.Context, clientUser []string, data map[string]any) error {
	var d changeGroupDescriptionT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.ChangeGroupDescription(ctx, d.GroupChatId, clientUser, d.NewDescription)

}

func changeGroupPicture(ctx context.Context, clientUser []string, data map[string]any) error {
	var d changeGroupPictureT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.ChangeGroupPicture(ctx, d.GroupChatId, clientUser, d.NewPictureData)
}

func addUsersToGroup(ctx context.Context, clientUser []string, data map[string]any) error {
	var d addUsersToGroupT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.AddUsersToGroup(ctx, d.GroupChatId, clientUser, d.NewUsers)
}

func removeUserFromGroup(ctx context.Context, clientUser []string, data map[string]any) error {
	var d actOnSingleUserT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.RemoveUserFromGroup(ctx, d.GroupChatId, clientUser, d.User)
}

func joinGroup(ctx context.Context, clientUser []string, data map[string]any) error {
	var d joinLeaveGroupT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.JoinGroup(ctx, d.GroupChatId, clientUser)
}

func leaveGroup(ctx context.Context, clientUser []string, data map[string]any) error {
	var d joinLeaveGroupT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.LeaveGroup(ctx, d.GroupChatId, clientUser)
}

func makeUserGroupAdmin(ctx context.Context, clientUser []string, data map[string]any) error {
	var d actOnSingleUserT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.MakeUserGroupAdmin(ctx, d.GroupChatId, clientUser, d.User)
}

func removeUserFromGroupAdmins(ctx context.Context, clientUser []string, data map[string]any) error {
	var d actOnSingleUserT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.RemoveUserFromGroupAdmins(ctx, d.GroupChatId, clientUser, d.User)
}
