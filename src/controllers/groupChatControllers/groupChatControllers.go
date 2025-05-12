package groupChatControllers

import (
	"context"
	"i9chat/src/appTypes"
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

	type action string

	type handler func(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error)

	actionToHandlerMap := map[action]handler{
		"join":                    joinGroup,
		"make user admin":         makeUserGroupAdmin,
		"add users":               addUsersToGroup,
		"change name":             changeGroupName,
		"change description":      changeGroupDescription,
		"change picture":          changeGroupPicture,
		"remove user from admins": removeUserFromGroupAdmins,
		"remove user":             removeUserFromGroup,
		"leave":                   leaveGroup,
	}

	var actionData map[string]any

	ad_err := c.BodyParser(&actionData)
	if ad_err != nil {
		return ad_err
	}

	actionHandler, ok := actionToHandlerMap[action(c.Params("action"))]
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "invalid group action")
	}

	respData, app_err := actionHandler(ctx, clientUser.Username, c.Params("group_id"), actionData)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
