package userService

import (
	"fmt"
	"i9chat/appTypes"
	dmChat "i9chat/models/chatModel/dmChatModel"
	groupChat "i9chat/models/chatModel/groupChatModel"
	user "i9chat/models/userModel"
	"i9chat/services/chatServices/dmChatService"
	"i9chat/services/chatServices/groupChatService"
	"i9chat/services/cloudStorageService"
	"i9chat/services/messageBrokerService"
	"log"
	"time"
)

func GoOnline(clientUserId int, userPOId string, mailbox chan<- any) error {
	userDMChatPartnersIdList, err := user.ChangePresence(clientUserId, "online", time.Now())
	if err != nil {
		return err
	}

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

func GoOffline(clientUserId int, lastSeen time.Time, userPOId string) {
	userDMChatPartnersIdList, err := user.ChangePresence(clientUserId, "offline", lastSeen)
	if err != nil {
		return
	}

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
}

func ChangeProfilePicture(clientUserId int, clientUsername string, pictureData []byte) (any, error) {
	// if any picture size validation error, do it here

	picPath := fmt.Sprintf("profile_pictures/%s/profile_pic_%d.jpg", clientUsername, time.Now().UnixNano())

	newPicUrl, err := cloudStorageService.UploadFile(picPath, pictureData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, ed_err := user.EditProfile(clientUserId, [][]string{{"profile_picture_url", newPicUrl}})
	if ed_err != nil {
		return nil, ed_err
	}

	respData := map[string]any{
		"msg": "Operation Successful",
	}

	return respData, nil
}

func UpdateMyLocation(clientUserId int, newGeolocation string) (any, error) {
	err := user.UpdateLocation(clientUserId, newGeolocation)
	if err != nil {
		return nil, err
	}

	respData := map[string]any{
		"msg": "Operation Successful",
	}

	return respData, err
}

func GetAllUsers(clientUserId int) ([]*user.User, error) {
	return user.GetAll(clientUserId)
}

func SearchUser(clientUserId int, query string) ([]*user.User, error) {
	return user.Search(clientUserId, query)
}

func FindNearbyUsers(clientUserId int, liveLocation string) ([]*user.User, error) {
	return user.FindNearby(clientUserId, liveLocation)
}

func GetMyChats(clientUserId int) ([]*map[string]any, error) {
	return user.GetChats(clientUserId)
}

func NewDMChat(initiatorId, partnerId int, initMsgContent map[string]any, createdAt time.Time) (*dmChat.InitiatorData, error) {
	return dmChatService.NewDMChat(initiatorId, partnerId, initMsgContent, createdAt)
}

func UpdateDMChatMessageDeliveryStatus(dmChatId, msgId, senderId, receiverId int, status string, updatedAt time.Time) {
	dmChatService.UpdateMessageDeliveryStatus(dmChatId, msgId, senderId, receiverId, status, updatedAt)
}

func BatchUpdateDMChatMessageDeliveryStatus(receiverId int, status string, ackDatas []*appTypes.DMChatMsgAckData) {
	dmChatService.BatchUpdateMessageDeliveryStatus(receiverId, status, ackDatas)
}

func NewGroupChat(name string, description string, pictureData []byte, creator []string, initUsers [][]appTypes.String) (*groupChat.CreatorData, error) {
	return groupChatService.NewGroupChat(name, description, pictureData, creator, initUsers)
}

func BatchUpdateGroupChatMessageDeliveryStatus(groupChatId, receiverId int, status string, ackDatas []*appTypes.GroupChatMsgAckData) {
	groupChatService.BatchUpdateMessageDeliveryStatus(groupChatId, receiverId, status, ackDatas)
}
