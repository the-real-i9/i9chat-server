package userService

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	user "i9chat/models/userModel"
	"i9chat/services/cloudStorageService"
	"i9chat/services/messageBrokerService"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

func GoOnline(ctx context.Context, clientUsername string) error {
	userDMChatPartnersIdList, err := user.ChangePresence(ctx, clientUsername, "online", time.Time{})
	if err != nil {
		return err
	}

	// work in progress
	// "recepients" are: all users to whom you are a DMChat partner
	go func(recepientIds []*int) {
		for _, rId := range recepientIds {
			rId := *rId

			messageBrokerService.Send(fmt.Sprintf("user-%d-topic", rId), messageBrokerService.Message{
				Event: "user presence changed",
				Data: map[string]any{
					"username": clientUsername,
					"presence": "online",
					"lastSeen": nil,
				},
			})
		}
	}(userDMChatPartnersIdList)

	return nil
}

func GoOffline(ctx context.Context, clientUsername string, lastSeen time.Time) {
	userDMChatPartnersIdList, err := user.ChangePresence(ctx, clientUsername, "offline", lastSeen)
	if err != nil {
		return
	}

	// "recepients" are: all users to whom you are a DMChat partner
	go func(recepientIds []*int) {
		for _, rId := range recepientIds {
			rId := *rId

			messageBrokerService.Send(fmt.Sprintf("user-%d-topic", rId), messageBrokerService.Message{
				Event: "user presence changed",
				Data: map[string]any{
					"username": clientUsername,
					"presence": "offline",
					"lastSeen": lastSeen,
				},
			})
		}
	}(userDMChatPartnersIdList)
}

func ChangeProfilePicture(ctx context.Context, clientUsername string, pictureData []byte) (any, error) {
	// if any picture size validation error, do it here

	ext := mimetype.Detect(pictureData).Extension()
	picPath := fmt.Sprintf("profile_pictures/%s/profile_pic_%d.%s", clientUsername, time.Now().UnixNano(), ext)

	newPicUrl, err := cloudStorageService.Upload(ctx, picPath, pictureData)
	if err != nil {
		return nil, err
	}

	ed_err := user.EditProfile(ctx, clientUsername, map[string]any{"profile_picture_url": newPicUrl})
	if ed_err != nil {
		return nil, ed_err
	}

	respData := map[string]any{
		"msg": "Operation Successful",
	}

	return respData, nil
}

func UpdateMyLocation(ctx context.Context, clientUsername string, newGeolocation *appTypes.UserGeolocation) (any, error) {
	err := user.UpdateLocation(ctx, clientUsername, newGeolocation)
	if err != nil {
		return nil, err
	}

	respData := map[string]any{
		"msg": "Operation Successful",
	}

	return respData, err
}

func SearchUser(ctx context.Context, clientUsername, query string) ([]any, error) {
	return user.Search(ctx, clientUsername, query)
}

func FindNearbyUsers(ctx context.Context, clientUsername string, long, lat, radius float64) ([]any, error) {
	return user.FindNearby(ctx, clientUsername, long, lat, radius)
}

func GetMyChats(ctx context.Context, clientUsername string) ([]any, error) {
	return user.GetChats(ctx, clientUsername)
}
