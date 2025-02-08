package groupChatControllers

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/services/chatServices/groupChatService"

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
		clientUser.Username,
		body.Name,
		body.Description,
		body.PictureData,
		body.InitUsers,
		body.CreatedAt,
	)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetChatHistory(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	groupChatId := c.Params("group_chat_id")

	var query getChatHistoryQuery

	query_err := c.QueryParser(&query)
	if query_err != nil {
		return query_err
	}

	if val_err := query.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, app_err := groupChatService.GetChatHistory(ctx, groupChatId, query.Limit, query.Offset)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)

}

func ExecuteAction(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	type handler func(ctx context.Context, clientUsername string, data map[string]any) error

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

	var body executeActionBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	app_err := actionToHandlerMap[body.Action](ctx, []string{fmt.Sprint(clientUser.Id), clientUser.Id}, body.Data)
	if app_err != nil {
		return app_err
	}

	return c.JSON(fiber.Map{
		"msg": "Operation Successful!",
	})
}

func changeGroupName(ctx context.Context, clientUsername string, data map[string]any) error {
	var d changeGroupNameAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.ChangeGroupName(ctx, d.GroupId, clientUsername, d.NewName)
}

func changeGroupDescription(ctx context.Context, clientUsername string, data map[string]any) error {
	var d changeGroupDescriptionAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.ChangeGroupDescription(ctx, d.GroupId, clientUsername, d.NewDescription)

}

func changeGroupPicture(ctx context.Context, clientUsername string, data map[string]any) error {
	var d changeGroupPictureAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.ChangeGroupPicture(ctx, d.GroupId, clientUsername, d.NewPictureData)
}

func addUsersToGroup(ctx context.Context, clientUsername string, data map[string]any) error {
	var d addUsersToGroupAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.AddUsersToGroup(ctx, d.GroupId, clientUsername, d.NewUsers)
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
