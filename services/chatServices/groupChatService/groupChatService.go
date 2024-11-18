package groupChatService

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	groupChat "i9chat/models/chatModel/groupChatModel"
	"i9chat/services/appServices"
	"i9chat/services/cloudStorageService"
	"i9chat/services/messageBrokerService"
	"time"
)

func NewGroupChat(name string, description string, pictureData []byte, creator []string, initUsers [][]appTypes.String) (*groupChat.CreatorData, error) {
	picUrl, _ := uploadGroupPicture(pictureData)

	newGroupChat, err := groupChat.New(name, description, picUrl, creator, initUsers)
	if err != nil {
		return nil, err
	}

	go broadcastNewGroup(initUsers, newGroupChat.InitMemberData)

	return newGroupChat.CreatorData, nil
}

func broadcastNewGroup(initMembers [][]appTypes.String, initMemberData *groupChat.InitMemberData) {
	for _, initMember := range initMembers {
		initMemberId := initMember[0]

		go messageBrokerService.PostMessage(fmt.Sprintf("user-%s", initMemberId), messageBrokerService.Message{
			Event: "new group chat",
			Data:  initMemberData,
		})
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
		go broadcastGroupChatMessageDeliveryStatusUpdate(groupChatId, receiverId, ackDatas, batchStatusUpdateResult.OverallDeliveryStatus)
	}
}

func broadcastGroupChatMessageDeliveryStatusUpdate(groupChatId, clientUserId int, ackDatas []*appTypes.GroupChatMsgAckData, status string) {
	membersIds, err := helpers.QueryRowsField[int]("SELECT member_id FROM group_chat_membership WHERE group_chat_id = $1 AND member_id != $2 AND deleted = false", groupChatId, clientUserId)
	if err == nil {
		for _, memberId := range membersIds {
			memberId := *memberId
			for _, data := range ackDatas {
				msgId := data.MsgId

				go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", memberId), messageBrokerService.Message{
					Event: "group chat message delivery status changed",
					Data: map[string]any{
						"groupChatId": groupChatId,
						"msgId":       msgId,
						"status":      status,
					},
				})
			}
		}
	}
}

func GetChatHistory(dmChatId, offset int) ([]*groupChat.HistoryItem, error) {
	return groupChat.GetChatHistory(dmChatId, offset)
}

func SendMessage(groupChatId, clientUserId int, msgContent map[string]any, createdAt time.Time) (*groupChat.SenderData, error) {

	modMsgContent := appServices.UploadMessageMedia(clientUserId, msgContent)

	newMessage, err := groupChat.SendMessage(groupChatId, clientUserId, modMsgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go broadcastNewMessage(newMessage.MembersIds, newMessage.MemberData)

	respData := newMessage.SenderData

	return respData, nil
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

func ChangeGroupPicture(groupChatId int, admin []string, newPictureData []byte) error {
	newPicUrl, err := uploadGroupPicture(newPictureData)
	if err != nil {
		return err
	}

	newActivity, err := groupChat.ChangePicture(groupChatId, admin, newPicUrl)
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

func AddUsersToGroup(groupChatId int, admin []string, newUsers [][]appTypes.String) error {
	newActivity, err := groupChat.AddUsers(groupChatId, admin, newUsers)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func RemoveUserFromGroup(groupChatId int, admin []string, user []appTypes.String) error {
	newActivity, err := groupChat.RemoveUser(groupChatId, admin, user)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func JoinGroup(groupChatId int, newUser []string) error {
	newActivity, err := groupChat.Join(groupChatId, newUser)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func LeaveGroup(groupChatId int, user []string) error {
	newActivity, err := groupChat.Leave(groupChatId, user)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func MakeUserGroupAdmin(groupChatId int, admin []string, user []appTypes.String) error {
	newActivity, err := groupChat.MakeUserAdmin(groupChatId, admin, user)
	if err != nil {
		return err
	}

	go broadcastActivity(newActivity)

	return nil
}

func RemoveUserFromGroupAdmins(groupChatId int, admin []string, user []appTypes.String) error {
	newActivity, err := groupChat.RemoveUserFromAdmins(groupChatId, admin, user)
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
