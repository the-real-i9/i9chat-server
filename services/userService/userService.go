package userService

import (
	"context"
	"fmt"
	user "i9chat/models/userModel"
	"i9chat/services/cloudStorageService"
	"i9chat/services/messageBrokerService"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

func GoOnline(ctx context.Context, clientUserId int) error {
	userDMChatPartnersIdList, err := user.ChangePresence(ctx, clientUserId, "online", time.Now())
	if err != nil {
		return err
	}

	// "recepients" are: all users to whom you are a DMChat partner
	go func(recepientIds []*int) {
		for _, rId := range recepientIds {
			rId := *rId

			messageBrokerService.Send(fmt.Sprintf("user-%d-topic", rId), messageBrokerService.Message{
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

func GoOffline(ctx context.Context, clientUserId int, lastSeen time.Time) {
	userDMChatPartnersIdList, err := user.ChangePresence(ctx, clientUserId, "offline", lastSeen)
	if err != nil {
		return
	}

	// "recepients" are: all users to whom you are a DMChat partner
	go func(recepientIds []*int) {
		for _, rId := range recepientIds {
			rId := *rId

			go messageBrokerService.Send(fmt.Sprintf("user-%d-topic", rId), messageBrokerService.Message{
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

func ChangeProfilePicture(ctx context.Context, clientUserId int, clientUsername string, pictureData []byte) (any, error) {
	// if any picture size validation error, do it here

	ext := mimetype.Detect(pictureData).Extension()
	picPath := fmt.Sprintf("profile_pictures/%s/profile_pic_%d.%s", clientUsername, time.Now().UnixNano(), ext)

	newPicUrl, err := cloudStorageService.Upload(ctx, picPath, pictureData)
	if err != nil {
		return nil, err
	}

	_, ed_err := user.EditProfile(ctx, clientUserId, [][]string{{"profile_picture_url", newPicUrl}})
	if ed_err != nil {
		return nil, ed_err
	}

	respData := map[string]any{
		"msg": "Operation Successful",
	}

	return respData, nil
}

func UpdateMyLocation(ctx context.Context, clientUserId int, newGeolocation string) (any, error) {
	err := user.UpdateLocation(ctx, clientUserId, newGeolocation)
	if err != nil {
		return nil, err
	}

	respData := map[string]any{
		"msg": "Operation Successful",
	}

	return respData, err
}

func GetAllUsers(ctx context.Context, clientUserId int) ([]*user.User, error) {
	return user.GetAll(ctx, clientUserId)
}

func SearchUser(ctx context.Context, clientUserId int, query string) ([]*user.User, error) {
	return user.Search(ctx, clientUserId, query)
}

func FindNearbyUsers(ctx context.Context, clientUserId int, liveLocation string) ([]*user.User, error) {
	return user.FindNearby(ctx, clientUserId, liveLocation)
}

func GetMyChats(ctx context.Context, clientUserId int) ([]*map[string]any, error) {
	return user.GetChats(ctx, clientUserId)
}
