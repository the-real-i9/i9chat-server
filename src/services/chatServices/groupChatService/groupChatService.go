package groupChatService

import (
	"context"
	"encoding/json"
	"fmt"
	"i9chat/src/appTypes"
	groupChat "i9chat/src/models/chatModel/groupChatModel"
	"i9chat/src/services/appServices"
	"i9chat/src/services/cloudStorageService"
	"i9chat/src/services/eventStreamService"
	"time"
)

func NewGroupChat(ctx context.Context, clientUsername, name, description string, pictureData []byte, initUsers []string, createdAt time.Time) (map[string]any, error) {
	picUrl, err := uploadGroupPicture(ctx, pictureData)
	if err != nil {
		return nil, err
	}

	newGroupChat, err := groupChat.New(ctx, clientUsername, name, description, picUrl, initUsers, createdAt)
	if err != nil {
		return nil, err
	}

	go broadcastNewGroup(initUsers, newGroupChat.InitMemberData)

	return newGroupChat.ClientData, nil
}

func GetChatHistory(ctx context.Context, clientUsername, groupId string, limit int, offset int64) (any, error) {
	return groupChat.ChatHistory(ctx, clientUsername, groupId, limit, time.UnixMilli(offset))
}

func GetGroupInfo(ctx context.Context, groupId string) (map[string]any, error) {
	return groupChat.GroupInfo(ctx, groupId)
}

func GetGroupMemInfo(ctx context.Context, clientUsername, groupId string) (map[string]any, error) {
	return groupChat.GroupMemInfo(ctx, clientUsername, groupId)
}

func SendMessage(ctx context.Context, groupId, clientUsername string, msgContent *appTypes.MsgContent, createdAt time.Time) (map[string]any, error) {

	err := appServices.UploadMessageMedia(ctx, clientUsername, msgContent)
	if err != nil {
		return nil, err
	}

	msgContentJson, _ := json.Marshal(*msgContent)

	newMessage, err := groupChat.SendMessage(ctx, groupId, clientUsername, msgContentJson, createdAt)
	if err != nil {
		return nil, err
	}

	go broadcastNewMessage(newMessage.MemberUsernames, newMessage.MemberData, groupId)

	return newMessage.ClientData, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, groupId, msgId string, deliveredAt time.Time) error {
	msgAck, err := groupChat.AckMessageDelivered(ctx, clientUsername, groupId, msgId, deliveredAt)
	if err != nil {
		return err
	}

	if msgAck.All {
		go broadcastMsgDelivered(msgAck.MemberUsernames, map[string]any{
			"group_id": groupId,
			"msg_id":   msgId,
		})
	}

	return nil
}

func AckMessageRead(ctx context.Context, clientUsername, groupId, msgId string, readAt time.Time) error {
	msgAck, err := groupChat.AckMessageRead(ctx, clientUsername, groupId, msgId, readAt)
	if err != nil {
		return err
	}

	if msgAck.All {
		go broadcastMsgRead(msgAck.MemberUsernames, map[string]any{
			"group_id": groupId,
			"msg_id":   msgId,
		})
	}

	return nil
}

func ChangeGroupName(ctx context.Context, groupId, clientUsername, newName string) (any, error) {
	newActivity, err := groupChat.ChangeName(ctx, groupId, clientUsername, newName)
	if err != nil {
		return nil, err
	}

	go broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)

	return newActivity.ClientData, nil
}

func ChangeGroupDescription(ctx context.Context, groupId, clientUsername, newDescription string) (any, error) {
	newActivity, err := groupChat.ChangeDescription(ctx, groupId, clientUsername, newDescription)
	if err != nil {
		return nil, err
	}

	go broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)

	return newActivity.ClientData, nil
}

func ChangeGroupPicture(ctx context.Context, groupId, clientUsername string, newPictureData []byte) (any, error) {
	newPicUrl, err := uploadGroupPicture(ctx, newPictureData)
	if err != nil {
		return nil, err
	}

	newActivity, err := groupChat.ChangePicture(ctx, groupId, clientUsername, newPicUrl)
	if err != nil {
		return nil, err
	}

	go broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)

	return newActivity.ClientData, nil
}

func uploadGroupPicture(ctx context.Context, pictureData []byte) (string, error) {
	if len(pictureData) < 1 {
		return "", fmt.Errorf("upload error: no picture data")
	}
	picPath := fmt.Sprintf("group_chat_pictures/group_chat_pic_%d.jpg", time.Now().UnixNano())

	picUrl, err := cloudStorageService.Upload(ctx, picPath, pictureData)

	if err != nil {
		return "", err
	}

	return picUrl, nil
}

func AddUsersToGroup(ctx context.Context, groupId, clientUsername string, newUsers []string) (any, error) {
	newActivity, newUserData, err := groupChat.AddUsers(ctx, groupId, clientUsername, newUsers)
	if err != nil {
		return nil, err
	}

	// a user with already existing group chat must check
	go func() {
		broadcastNewGroup(newUsers, newUserData)

		broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
	}()

	return newActivity.ClientData, nil
}

func RemoveUserFromGroup(ctx context.Context, groupId, clientUsername, targetUser string) (any, error) {
	newActivity, targetUserData, err := groupChat.RemoveUser(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return nil, err
	}

	go func() {
		eventStreamService.Send(targetUser, appTypes.ServerWSMsg{
			Event: "removed from group",
			Data: map[string]any{
				"group_id": groupId,
			},
		})

		broadcastActivity([]string{targetUser}, targetUserData, groupId)

		broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
	}()

	return newActivity.ClientData, nil
}

func JoinGroup(ctx context.Context, groupId, clientUsername string) (any, error) {
	newActivity, err := groupChat.Join(ctx, groupId, clientUsername)
	if err != nil {
		return nil, err
	}

	go broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)

	return newActivity.ClientData, nil
}

func LeaveGroup(ctx context.Context, groupId, clientUsername string) (any, error) {
	newActivity, err := groupChat.Leave(ctx, groupId, clientUsername)
	if err != nil {
		return nil, err
	}

	go broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)

	return newActivity.ClientData, nil
}

func MakeUserGroupAdmin(ctx context.Context, groupId, clientUsername, targetUser string) (any, error) {
	newActivity, newAdminData, err := groupChat.MakeUserAdmin(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return nil, err
	}

	go func() {
		broadcastActivity([]string{targetUser}, newAdminData, groupId)

		broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
	}()

	return newActivity.ClientData, nil
}

func RemoveUserFromGroupAdmins(ctx context.Context, groupId, clientUsername, targetUser string) (any, error) {
	newActivity, oldAdminData, err := groupChat.RemoveUserFromAdmins(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return nil, err
	}

	go func() {
		broadcastActivity([]string{targetUser}, oldAdminData, groupId)

		broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
	}()

	return newActivity.ClientData, nil
}
