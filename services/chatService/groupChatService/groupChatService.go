package groupChatService

import (
	"fmt"
	groupChat "i9chat/models/chatModel/groupChatModel"
	"i9chat/services/appObservers"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"time"
)

func broadcastNewGroup(initMembers [][]appTypes.String, initMemberData *groupChat.InitMemberData) {
	for _, initMember := range initMembers {
		initMemberId := initMember[0]
		go appObservers.GroupChatObserver{}.Send(fmt.Sprintf("user-%s", initMemberId), initMemberData, "new chat")
	}
}

func uploadGroupPicture(pictureData []byte) (string, error) {
	if len(pictureData) < 1 {
		return "", fmt.Errorf("upload error: no picture data")
	}
	picPath := fmt.Sprintf("group_chat_pictures/group_chat_pic_%d.jpg", time.Now().UnixNano())

	picUrl, err := helpers.UploadFile(picPath, pictureData)

	if err != nil {
		return "", err
	}

	return picUrl, nil
}

func NewGroupChat(name string, description string, pictureData []byte, creator []string, initUsers [][]appTypes.String) (*groupChat.CreatorData, error) {
	picUrl, _ := uploadGroupPicture(pictureData)

	newGroupChat, err := groupChat.New(name, description, picUrl, creator, initUsers)
	if err != nil {
		return nil, err
	}

	go broadcastNewGroup(initUsers, newGroupChat.InitMemberData)

	return newGroupChat.CreatorData, nil
}

func broadcastActivity(newActivity *groupChat.NewActivity) {

	for _, mId := range newActivity.MembersIds {
		go appObservers.GroupChatObserver{}.Send(fmt.Sprintf("user-%d", mId), newActivity.ActivityInfo, "new activity")
	}
}

func ChangeGroupName(groupChatId int, clientUser []string, newName string) error {
	newActivity, err := groupChat.ChangeName(groupChatId, clientUser, newName)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func ChangeGroupDescription(groupChatId int, clientUser []string, newDescription string) error {
	newActivity, err := groupChat.ChangeDescription(groupChatId, clientUser, newDescription)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func ChangeGroupPicture(groupChatId int, clientUser []string, newPictureData []byte) error {
	newPicUrl, err := uploadGroupPicture(newPictureData)
	if err != nil {
		return err
	}

	newActivity, err := groupChat.ChangePicture(groupChatId, clientUser, newPicUrl)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func AddUsersToGroup(groupChatId int, clientUser []string, newUsers [][]appTypes.String) error {
	newActivity, err := groupChat.AddUsers(groupChatId, clientUser, newUsers)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func RemoveUserFromGroup(groupChatId int, clientUser []string, user []appTypes.String) error {
	newActivity, err := groupChat.RemoveUser(groupChatId, clientUser, user)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func JoinGroup(groupChatId int, clientUser []string) error {
	newActivity, err := groupChat.Join(groupChatId, clientUser)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func LeaveGroup(groupChatId int, clientUser []string) error {
	newActivity, err := groupChat.Leave(groupChatId, clientUser)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func MakeUserGroupAdmin(groupChatId int, clientUser []string, user []appTypes.String) error {
	newActivity, err := groupChat.MakeUserAdmin(groupChatId, clientUser, user)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func RemoveUserFromGroupAdmins(groupChatId int, clientUser []string, user []appTypes.String) error {
	newActivity, err := groupChat.RemoveUserFromAdmins(groupChatId, clientUser, user)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func broadcastNewMessage(membersIds []int, memberData *groupChat.MemberData) {
	for _, mId := range membersIds {
		memberId := mId
		go appObservers.GroupChatObserver{}.Send(fmt.Sprintf("user-%d", memberId), memberData, "new message")
	}
}

func SendMessage(groupChatId, clientUserId int, msgContent map[string]any, createdAt time.Time) (*groupChat.SenderData, error) {

	newMessage, err := groupChat.SendMessage(groupChatId, clientUserId, msgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go broadcastNewMessage(newMessage.MembersIds, newMessage.MemberData)

	return newMessage.SenderData, nil
}

func broadcastMessageDeliveryStatusUpdate(groupChatId, clientUserId int, ackDatas []*appTypes.GroupChatMsgAckData, status string) {
	membersIds, err := helpers.QueryRowsField[int]("SELECT member_id FROM group_chat_membership WHERE group_chat_id = $1 AND member_id != $2 AND deleted = false", groupChatId, clientUserId)
	if err == nil {
		for _, memberId := range membersIds {
			memberId := *memberId
			for _, data := range ackDatas {
				msgId := data.MsgId
				go appObservers.DMChatSessionObserver{}.Send(
					fmt.Sprintf("user-%d--groupchat-%d", memberId, groupChatId),
					map[string]any{"msgId": msgId, "status": status},
					"delivery status update",
				)
			}
		}
	}
}

func BatchUpdateMessageDeliveryStatus(groupChatId, receiverId int, status string, ackDatas []*appTypes.GroupChatMsgAckData) {
	batchStatusUpdateResult, err := groupChat.BatchUpdateMessageDeliveryStatus(groupChatId, receiverId, status, ackDatas)
	if err != nil {
		return
	}

	// The idea is that: the delivery status of a group message updates only
	// when all members have acknowledged the message as either "delivered" or "seen",
	// the new delivery status is set in the OverallDeliveryStatus, after all members have acknowledged a certain delivery status ("delivered" or "seen").

	// ShouldBroadcast sets the boolean that determines if we should broadcast the OverallDeliveryStatus
	// (if it has changed since the last one) or not (if it hasn't changed since the last one)
	// this prevents unnecessary data transfer

	if batchStatusUpdateResult.ShouldBroadcast {
		go broadcastMessageDeliveryStatusUpdate(groupChatId, receiverId, ackDatas, batchStatusUpdateResult.OverallDeliveryStatus)
	}
}
