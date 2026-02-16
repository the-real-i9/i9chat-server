package groupChatControllers

import (
	"context"
	"i9chat/src/helpers"
	"i9chat/src/services/chatServices/groupChatService"

	"github.com/vmihailenco/msgpack/v5"
)

func changeGroupName(ctx context.Context, clientUsername, groupId string, data msgpack.RawMessage) (any, error) {

	d := helpers.FromBtMsgPack[changeGroupNameAction](data)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.ChangeGroupName(ctx, groupId, clientUsername, d.NewName)
}

func changeGroupDescription(ctx context.Context, clientUsername, groupId string, data msgpack.RawMessage) (any, error) {

	d := helpers.FromBtMsgPack[changeGroupDescriptionAction](data)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.ChangeGroupDescription(ctx, groupId, clientUsername, d.NewDescription)

}

func changeGroupPicture(ctx context.Context, clientUsername, groupId string, data msgpack.RawMessage) (any, error) {

	d := helpers.FromBtMsgPack[changeGroupPictureAction](data)

	if err := d.Validate(ctx); err != nil {
		return nil, err
	}

	return groupChatService.ChangeGroupPicture(ctx, groupId, clientUsername, d.PictureCloudName)
}

func addUsersToGroup(ctx context.Context, clientUsername, groupId string, data msgpack.RawMessage) (any, error) {

	d := helpers.FromBtMsgPack[addUsersToGroupAction](data)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.AddUsersToGroup(ctx, groupId, clientUsername, d.NewUsers)
}

func removeUserFromGroup(ctx context.Context, clientUsername, groupId string, data msgpack.RawMessage) (any, error) {

	d := helpers.FromBtMsgPack[actOnSingleUserAction](data)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.RemoveUserFromGroup(ctx, groupId, clientUsername, d.User)
}

func joinGroup(ctx context.Context, clientUsername, groupId string, _ msgpack.RawMessage) (any, error) {
	return groupChatService.JoinGroup(ctx, groupId, clientUsername)
}

func leaveGroup(ctx context.Context, clientUsername, groupId string, _ msgpack.RawMessage) (any, error) {
	return groupChatService.LeaveGroup(ctx, groupId, clientUsername)
}

func makeUserGroupAdmin(ctx context.Context, clientUsername, groupId string, data msgpack.RawMessage) (any, error) {

	d := helpers.FromBtMsgPack[actOnSingleUserAction](data)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.MakeUserGroupAdmin(ctx, groupId, clientUsername, d.User)
}

func removeUserFromGroupAdmins(ctx context.Context, clientUsername, groupId string, data msgpack.RawMessage) (any, error) {
	d := helpers.FromBtMsgPack[actOnSingleUserAction](data)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.RemoveUserFromGroupAdmins(ctx, groupId, clientUsername, d.User)
}
