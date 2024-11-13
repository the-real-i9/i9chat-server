package userService

import (
	"fmt"
	user "i9chat/models/userModel"
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
