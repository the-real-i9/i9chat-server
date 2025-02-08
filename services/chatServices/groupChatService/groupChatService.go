package groupChatService

import (
	"context"
	"encoding/json"
	"fmt"
	"i9chat/appTypes"
	groupChat "i9chat/models/chatModel/groupChatModel"
	"i9chat/services/appServices"
	"i9chat/services/cloudStorageService"
	"i9chat/services/messageBrokerService"
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

func broadcastNewGroup(initMembers []string, initMemberData map[string]any) {
	for _, initMember := range initMembers {

		messageBrokerService.Send(fmt.Sprintf("user-%s-topic", initMember), messageBrokerService.Message{
			Event: "new group chat",
			Data:  initMemberData,
		})
	}
}

func GetChatHistory(ctx context.Context, groupChatId string, limit int, offset time.Time) ([]any, error) {
	return groupChat.GetChatHistory(ctx, groupChatId, limit, offset)
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

	go broadcastNewMessage(newMessage.MemberUsernames, newMessage.MemberData)

	return newMessage.ClientData, nil
}

func broadcastNewMessage(memberUsernames []string, memberData map[string]any) {
	for _, mu := range memberUsernames {

		messageBrokerService.Send(fmt.Sprintf("user-%s-topic", mu), messageBrokerService.Message{
			Event: "new group chat message",
			Data:  memberData,
		})
	}
}

// work in progress
func AckMessageDelivered(ctx context.Context, clientUsername, groupId, msgId string, deliveredAt time.Time) error {
	if err := groupChat.AckMessageDelivered(ctx, clientUsername, groupId, msgId, deliveredAt); err != nil {
		return err
	}

	// implement broadcast message delivered when appropriate

	return nil
}

// work in progress
func AckMessageRead(ctx context.Context, clientUsername, groupId, msgId string, readAt time.Time) error {
	if err := groupChat.AckMessageRead(ctx, clientUsername, groupId, msgId, readAt); err != nil {
		return err
	}

	// implement broadcast message read when appropriate

	return nil
}

func ChangeGroupName(ctx context.Context, groupId, clientUsername, newName string) error {
	newActivity, err := groupChat.ChangeName(ctx, groupId, clientUsername, newName)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func ChangeGroupDescription(ctx context.Context, groupChatId int, clientUser []string, newDescription string) error {
	newActivity, err := groupChat.ChangeDescription(ctx, groupChatId, clientUser, newDescription)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func ChangeGroupPicture(ctx context.Context, groupChatId int, admin []string, newPictureData []byte) error {
	newPicUrl, err := uploadGroupPicture(ctx, newPictureData)
	if err != nil {
		return err
	}

	newActivity, err := groupChat.ChangePicture(ctx, groupChatId, admin, newPicUrl)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
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

func AddUsersToGroup(ctx context.Context, groupChatId int, admin []string, newUsers [][]appTypes.String) error {
	newActivity, err := groupChat.AddUsers(ctx, groupChatId, admin, newUsers)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func RemoveUserFromGroup(ctx context.Context, groupChatId int, admin []string, user []appTypes.String) error {
	newActivity, err := groupChat.RemoveUser(ctx, groupChatId, admin, user)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func JoinGroup(ctx context.Context, groupChatId int, newUser []string) error {
	newActivity, err := groupChat.Join(ctx, groupChatId, newUser)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func LeaveGroup(ctx context.Context, groupChatId int, user []string) error {
	newActivity, err := groupChat.Leave(ctx, groupChatId, user)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func MakeUserGroupAdmin(ctx context.Context, groupChatId int, admin []string, user []appTypes.String) error {
	newActivity, err := groupChat.MakeUserAdmin(ctx, groupChatId, admin, user)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func RemoveUserFromGroupAdmins(ctx context.Context, groupChatId int, admin []string, user []appTypes.String) error {
	newActivity, err := groupChat.RemoveUserFromAdmins(ctx, groupChatId, admin, user)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func broadcastActivity(newActivity groupChat.NewActivity) {

	for _, mId := range newActivity.MembersIds {
		messageBrokerService.Send(fmt.Sprintf("user-%d-topic", mId), messageBrokerService.Message{
			Event: "new group chat activity",
			Data:  newActivity.ActivityInfo,
		})
	}
}
