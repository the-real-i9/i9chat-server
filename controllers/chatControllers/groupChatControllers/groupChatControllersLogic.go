package groupChatControllers

import (
	"fmt"
	"i9chat/helpers"
	groupChat "i9chat/models/chatModel/groupChatModel"
	"i9chat/services/cloudStorageService"
	"i9chat/services/messageBrokerService"
	"time"
)

func sendMessage(groupChatId, clientUserId int, msgContent map[string]any, createdAt time.Time) (*groupChat.SenderData, error) {

	newMessage, err := groupChat.SendMessage(groupChatId, clientUserId, msgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go broadcastNewMessage(newMessage.MembersIds, newMessage.MemberData)

	return newMessage.SenderData, nil
}

func broadcastNewMessage(membersIds []int, memberData *groupChat.MemberData) {
	for _, mId := range membersIds {
		memberId := mId
		go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", memberId), messageBrokerService.Message{
			Event: "new group chat message",
			Data:  memberData,
		})
	}
}

func changeGroupName(clientUser []string, data map[string]any) error {
	var d changeGroupNameT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	newActivity, err := groupChat.ChangeName(d.GroupChatId, clientUser, d.NewName)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func changeGroupDescription(clientUser []string, data map[string]any) error {
	var d changeGroupDescriptionT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	newActivity, err := groupChat.ChangeDescription(d.GroupChatId, clientUser, d.NewDescription)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func changeGroupPicture(clientUser []string, data map[string]any) error {
	var d changeGroupPictureT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	newPicUrl, err := uploadGroupPicture(d.NewPictureData)
	if err != nil {
		return err
	}

	newActivity, err := groupChat.ChangePicture(d.GroupChatId, clientUser, newPicUrl)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func uploadGroupPicture(pictureData []byte) (string, error) {
	if len(pictureData) < 1 {
		return "", fmt.Errorf("upload error: no picture data")
	}
	picPath := fmt.Sprintf("group_chat_pictures/group_chat_pic_%d.jpg", time.Now().UnixNano())

	picUrl, err := cloudStorageService.UploadFile(picPath, pictureData)

	if err != nil {
		return "", err
	}

	return picUrl, nil
}

func addUsersToGroup(clientUser []string, data map[string]any) error {
	var d addUsersToGroupT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	newActivity, err := groupChat.AddUsers(d.GroupChatId, clientUser, d.NewUsers)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func removeUserFromGroup(clientUser []string, data map[string]any) error {
	var d actOnSingleUserT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	newActivity, err := groupChat.RemoveUser(d.GroupChatId, clientUser, d.User)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func joinGroup(clientUser []string, data map[string]any) error {
	var d joinLeaveGroupT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	newActivity, err := groupChat.Join(d.GroupChatId, clientUser)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func leaveGroup(clientUser []string, data map[string]any) error {
	var d joinLeaveGroupT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	newActivity, err := groupChat.Leave(d.GroupChatId, clientUser)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func makeUserGroupAdmin(clientUser []string, data map[string]any) error {
	var d actOnSingleUserT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	newActivity, err := groupChat.MakeUserAdmin(d.GroupChatId, clientUser, d.User)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func removeUserFromGroupAdmins(clientUser []string, data map[string]any) error {
	var d actOnSingleUserT

	helpers.MapToStruct(data, &d)

	if val_err := d.Validate(); val_err != nil {
		return val_err
	}

	newActivity, err := groupChat.RemoveUserFromAdmins(d.GroupChatId, clientUser, d.User)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func broadcastActivity(newActivity *groupChat.NewActivity) {

	for _, mId := range newActivity.MembersIds {
		go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", mId), messageBrokerService.Message{
			Event: "new group chat activity",
			Data:  newActivity.ActivityInfo,
		})
	}
}
