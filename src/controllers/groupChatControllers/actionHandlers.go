package groupChatControllers

import (
	"context"
	"i9chat/src/helpers"
	"i9chat/src/services/chatServices/groupChatService"
)

func changeGroupName(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d changeGroupNameAction

	helpers.ToStruct(data, &d)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.ChangeGroupName(ctx, groupId, clientUsername, d.NewName)
}

func changeGroupDescription(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d changeGroupDescriptionAction

	helpers.ToStruct(data, &d)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.ChangeGroupDescription(ctx, groupId, clientUsername, d.NewDescription)

}

func changeGroupPicture(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d changeGroupPictureAction

	helpers.ToStruct(data, &d)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.ChangeGroupPicture(ctx, groupId, clientUsername, d.NewPictureData)
}

func addUsersToGroup(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d addUsersToGroupAction

	helpers.ToStruct(data, &d)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.AddUsersToGroup(ctx, groupId, clientUsername, d.NewUsers)
}

func removeUserFromGroup(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d actOnSingleUserAction

	helpers.ToStruct(data, &d)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.RemoveUserFromGroup(ctx, groupId, clientUsername, d.User)
}

func joinGroup(ctx context.Context, clientUsername, groupId string, _ map[string]any) (any, error) {
	return groupChatService.JoinGroup(ctx, groupId, clientUsername)
}

func leaveGroup(ctx context.Context, clientUsername, groupId string, _ map[string]any) (any, error) {
	return groupChatService.LeaveGroup(ctx, groupId, clientUsername)
}

func makeUserGroupAdmin(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d actOnSingleUserAction

	helpers.ToStruct(data, &d)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.MakeUserGroupAdmin(ctx, groupId, clientUsername, d.User)
}

func removeUserFromGroupAdmins(ctx context.Context, clientUsername, groupId string, data map[string]any) (any, error) {
	var d actOnSingleUserAction

	helpers.ToStruct(data, &d)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.RemoveUserFromGroupAdmins(ctx, groupId, clientUsername, d.User)
}
