package userControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	dmChat "i9chat/models/chatModel/dmChatModel"
	groupChat "i9chat/models/chatModel/groupChatModel"
	user "i9chat/models/userModel"
	"i9chat/services/cloudStorageService"
	"i9chat/services/messageBrokerService"
	"log"
	"time"
)

func goOnline(clientUserId int, userPOId string, mailbox chan<- any) error {
	userDMChatPartnersIdList := user.ChangePresence(clientUserId, "online", time.Now())

	go messageBrokerService.AddMailbox(userPOId, mailbox)

	// "recepients" are: all users to whom you are a DMChat partner
	go func(recepientIds []*int) {
		for _, rId := range recepientIds {
			rId := *rId

			go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", rId), messageBrokerService.Message{
				Event: "user presence changed",
				Data: map[string]any{
					"userId":   clientUserId,
					"presence": "online",
					"lastSeen": nil,
				},
			})
		}
	}(userDMChatPartnersIdList)

	return nil
}

func goOffline(clientUserId int, lastSeen time.Time, userPOId string) error {
	userDMChatPartnersIdList := user.ChangePresence(clientUserId, "offline", lastSeen)

	go messageBrokerService.RemoveMailbox(userPOId)

	// "recepients" are: all users to whom you are a DMChat partner
	go func(recepientIds []*int) {
		for _, rId := range recepientIds {
			rId := *rId

			go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", rId), messageBrokerService.Message{
				Event: "user presence changed",
				Data: map[string]any{
					"userId":   clientUserId,
					"presence": "offline",
					"lastSeen": lastSeen,
				},
			})
		}
	}(userDMChatPartnersIdList)

	return nil
}

func changeMyProfilePicture(clientUserId int, clientUsername string, pictureData []byte) error {
	// if any picture size validation error, do it here

	picPath := fmt.Sprintf("profile_pictures/%s/profile_pic_%d.jpg", clientUsername, time.Now().UnixNano())

	newPicUrl, err := cloudStorageService.UploadFile(picPath, pictureData)
	if err != nil {
		log.Println(err)
		return err
	}

	_, ed_err := user.EditProfile(clientUserId, [][]string{{"profile_picture_url", newPicUrl}})
	if ed_err != nil {
		return ed_err
	}

	return nil
}

func newDMChat(initiatorId, partnerId int, initMsgContent map[string]any, createdAt time.Time) (*dmChat.InitiatorData, error) {
	dmChat, err := dmChat.New(initiatorId, partnerId, initMsgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", partnerId), messageBrokerService.Message{
		Event: "new dm chat",
		Data:  dmChat.PartnerData,
	})

	return dmChat.InitiatorData, nil
}

func updateDMChatMessageDeliveryStatus(dmChatId, msgId, senderId, receiverId int, status string, updatedAt time.Time) {
	if err := dmChat.UpdateMessageDeliveryStatus(dmChatId, msgId, receiverId, status, updatedAt); err == nil {

		go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", senderId), messageBrokerService.Message{
			Event: "dm chat message delivery status changed",
			Data: map[string]any{
				"dmChatId": dmChatId,
				"msgId":    msgId,
				"status":   status,
			},
		})
	}
}

func batchUpdateDMChatMessageDeliveryStatus(receiverId int, status string, ackDatas []*appTypes.DMChatMsgAckData) {
	if err := dmChat.BatchUpdateMessageDeliveryStatus(receiverId, status, ackDatas); err == nil {
		for _, data := range ackDatas {
			data := data

			go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", data.SenderId), messageBrokerService.Message{
				Event: "dm chat message delivery status changed",
				Data: map[string]any{
					"dmChatId": data.DMChatId,
					"msgId":    data.MsgId,
					"status":   status,
				},
			})
		}
	}
}

func newGroupChat(name string, description string, pictureData []byte, creator []string, initUsers [][]appTypes.String) (*groupChat.CreatorData, error) {
	picUrl, _ := uploadGroupPicture(pictureData)

	newGroupChat, err := groupChat.New(name, description, picUrl, creator, initUsers)
	if err != nil {
		return nil, err
	}

	go broadcastNewGroup(initUsers, newGroupChat.InitMemberData)

	return newGroupChat.CreatorData, nil
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

func batchUpdateGroupChatMessageDeliveryStatus(groupChatId, receiverId int, status string, ackDatas []*appTypes.GroupChatMsgAckData) {
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

func broadcastNewGroup(initMembers [][]appTypes.String, initMemberData *groupChat.InitMemberData) {
	for _, initMember := range initMembers {
		initMemberId := initMember[0]

		go messageBrokerService.PostMessage(fmt.Sprintf("user-%s", initMemberId), messageBrokerService.Message{
			Event: "new group chat",
			Data:  initMemberData,
		})
	}
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
