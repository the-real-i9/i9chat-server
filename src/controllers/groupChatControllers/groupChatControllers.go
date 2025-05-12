package groupChatControllers

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/chatServices/groupChatService"

	"github.com/gofiber/fiber/v2"
)

func CreateNewGroupChat(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body newGroupChatBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
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

	return c.Status(fiber.StatusCreated).JSON(respData)
}

func ExecuteAction(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	type handler func(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error)

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

	var params executeActionParams
	params_err := c.BodyParser(&params)
	if params_err != nil {
		return params_err
	}

	if val_err := params.Validate(); val_err != nil {
		return val_err
	}

	var actionData map[string]any

	ad_err := c.BodyParser(&actionData)
	if ad_err != nil {
		return ad_err
	}

	respData, app_err := actionToHandlerMap[params.Action](ctx, clientUser.Username, params.GroupId, actionData)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func changeGroupName(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d changeGroupNameAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.ChangeGroupName(ctx, groupId, clientUsername, d.NewName)
}

func changeGroupDescription(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d changeGroupDescriptionAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.ChangeGroupDescription(ctx, groupId, clientUsername, d.NewDescription)

}

func changeGroupPicture(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d changeGroupPictureAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.ChangeGroupPicture(ctx, groupId, clientUsername, d.NewPictureData)
}

func addUsersToGroup(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d addUsersToGroupAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.AddUsersToGroup(ctx, groupId, clientUsername, d.NewUsers)
}

func removeUserFromGroup(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d actOnSingleUserAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.RemoveUserFromGroup(ctx, groupId, clientUsername, d.User)
}

func joinGroup(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	return groupChatService.JoinGroup(ctx, groupId, clientUsername)
}

func leaveGroup(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	return groupChatService.LeaveGroup(ctx, groupId, clientUsername)
}

func makeUserGroupAdmin(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d actOnSingleUserAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.MakeUserGroupAdmin(ctx, groupId, clientUsername, d.User)
}

func removeUserFromGroupAdmins(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d actOnSingleUserAction

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.RemoveUserFromGroupAdmins(ctx, groupId, clientUsername, d.User)
}
