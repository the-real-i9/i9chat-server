package groupChatControllers

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/services/chatServices/groupChatService"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

func AuthorizeGroupPicUpload(c *fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeGroupPicUploadBody

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := groupChatService.AuthorizeGroupPicUpload(ctx, body.PicMIME, body.PicSize)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func CreateNewGroup(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body newGroupChatBody

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, err := groupChatService.NewGroup(ctx,
		clientUser.Username,
		body.Name,
		body.Description,
		body.PictureData,
		body.InitUsers,
		body.CreatedAt,
	)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(respData)
}

func GetGroupMembers(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := groupChatService.GetGroupMembers(ctx, clientUser.Username, c.Params("group_id"), c.QueryInt("limit", 100), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetGroupChatHistory(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := groupChatService.GetChatHistory(ctx, clientUser.Username, c.Params("group_id"), c.QueryInt("limit", 50), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func ExecuteAction(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	type handler func(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error)

	actionToHandlerMap := map[string]handler{
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

	action, err := url.PathUnescape(c.Params("action"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid action parameter: %v", err))
	}

	actionHandler, ok := actionToHandlerMap[action]
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid group action: %s", c.Params("action")))
	}

	respData, err := actionHandler(ctx, clientUser.Username, c.Params("group_id"), actionData)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}
