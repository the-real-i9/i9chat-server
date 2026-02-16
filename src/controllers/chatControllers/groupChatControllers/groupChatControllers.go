package groupChatControllers

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/chatServices/groupChatService"
	"net/url"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

func AuthorizeGroupPicUpload(c fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeGroupPicUploadBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := groupChatService.AuthorizeGroupPicUpload(ctx, body.PicMIME)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func CreateNewGroup(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body newGroupChatBody

	err := c.Bind().MsgPack(&body)
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
		body.PictureCloudName,
		body.InitUsers,
		body.CreatedAt,
	)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(respData)
}

func GetGroupMembers(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := groupChatService.GetGroupMembers(ctx, clientUser.Username, c.Params("group_id"), helpers.CoalesceInt(query.Limit, 100), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetGroupChatHistory(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := groupChatService.GetChatHistory(ctx, clientUser.Username, c.Params("group_id"), helpers.CoalesceInt(query.Limit, 50), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func ExecuteAction(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	type handler func(ctx context.Context, clientUsername, groupId string, data msgpack.RawMessage) (any, error)

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

	var actionData msgpack.RawMessage

	err := c.Bind().MsgPack(&actionData)
	if err != nil {
		return err
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

	return c.MsgPack(respData)
}

func SendMessage(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (map[string]any, error) {

	acd := helpers.FromBtMsgPack[sendGroupChatMsg](actionData)

	if err := acd.Validate(ctx); err != nil {
		return nil, err
	}

	return groupChatService.SendMessage(ctx, clientUsername, acd.GroupId, acd.ReplyTargetMsgId, acd.IsReply, helpers.ToJson(acd.Msg), acd.At)
}

func AckMessageDelivered(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {

	acd := helpers.FromBtMsgPack[groupChatMsgAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.AckMessageDelivered(ctx, clientUsername, acd.GroupId, acd.MsgIds, acd.At)
}

func AckMessageRead(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {

	acd := helpers.FromBtMsgPack[groupChatMsgAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.AckMessageRead(ctx, clientUsername, acd.GroupId, acd.MsgIds, acd.At)
}

func GetGroupInfo(ctx context.Context, actionData msgpack.RawMessage) (any, error) {

	acd := helpers.FromBtMsgPack[groupInfo](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.GetGroupInfo(ctx, acd.GroupId)
}
